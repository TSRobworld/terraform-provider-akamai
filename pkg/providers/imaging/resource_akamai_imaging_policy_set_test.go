package imaging

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/tj/assert"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/imaging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceImagingPolicySet(t *testing.T) {
	var (
		anError = errors.New("oops")

		expectPolicySetCreation = func(t *testing.T, client *mockimaging, contractID, name, region, mediaType string, policySet *imaging.PolicySet, createError error) {
			client.On("CreatePolicySet", mock.Anything, imaging.CreatePolicySetRequest{
				ContractID: contractID,
				CreatePolicySet: imaging.CreatePolicySet{
					Name:   name,
					Region: imaging.Region(region),
					Type:   imaging.MediaType(mediaType),
				},
			}).Return(policySet, createError).Once()
		}

		expectPolicySetRead = func(t *testing.T, client *mockimaging, contractID, policySetID string, policySet *imaging.PolicySet, getPolicyError error, times int) {
			client.On("GetPolicySet", mock.Anything, imaging.GetPolicySetRequest{
				PolicySetID: policySetID, ContractID: contractID,
			}).Return(policySet, getPolicyError).Times(times)
		}

		expectPolicySetUpdate = func(t *testing.T, client *mockimaging, contractID, policySetID, name, region string, updatePolicySetError error) {
			client.On("UpdatePolicySet", mock.Anything, imaging.UpdatePolicySetRequest{
				PolicySetID: policySetID,
				ContractID:  contractID,
				UpdatePolicySet: imaging.UpdatePolicySet{
					Name:   name,
					Region: imaging.Region(region),
				},
			}).Return(nil, updatePolicySetError).Once()
		}
		expectPolicySetDelete = func(t *testing.T, client *mockimaging, contractID, policySetID, network string, listPolicyResponse *imaging.ListPoliciesResponse, listPolicyError, deletePolicySetError error) {
			client.On("ListPolicies", mock.Anything, imaging.ListPoliciesRequest{
				Network:     imaging.PolicyNetworkProduction,
				ContractID:  contractID,
				PolicySetID: policySetID,
			}).Return(listPolicyResponse, listPolicyError).Once()

			if listPolicyError != nil {
				return
			}

			client.On("ListPolicies", mock.Anything, imaging.ListPoliciesRequest{
				Network:     imaging.PolicyNetworkStaging,
				ContractID:  contractID,
				PolicySetID: policySetID,
			}).Return(listPolicyResponse, listPolicyError).Once()

			client.On("DeletePolicySet", mock.Anything, imaging.DeletePolicySetRequest{
				PolicySetID: policySetID,
				ContractID:  contractID,
			}).Return(deletePolicySetError).Once()
		}
	)

	testDir := "testdata/TestResPolicySet"

	tests := map[string]struct {
		init  func(*mockimaging)
		steps []resource.TestStep
	}{
		"ok create": {
			init: func(m *mockimaging) {
				contractID, policySetID, policySetName, region, mediaType := "1-TEST", "testID", "test_policy_set", "EMEA", string(imaging.TypeImage)
				createdPolicySet := &imaging.PolicySet{Name: policySetName, ID: policySetID, Region: imaging.Region(region), Type: mediaType}

				// create
				expectPolicySetCreation(t, m, contractID, policySetName, region, mediaType, createdPolicySet, nil)

				expectPolicySetRead(t, m, contractID, policySetID, createdPolicySet, nil, 2)

				// delete
				expectPolicySetDelete(t, m, contractID, policySetID, "", &imaging.ListPoliciesResponse{
					ItemKind: "POLICY",
					Items: []imaging.PolicyOutput{
						&imaging.PolicyOutputImage{ID: ".auto"},
					},
					TotalItems: 1,
				}, nil, nil)
			},
			steps: []resource.TestStep{
				{
					Config: loadFixtureString(fmt.Sprintf("%s/lifecycle/create.tf", testDir)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "id", "testID"),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "name", "test_policy_set"),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "region", string(imaging.RegionEMEA)),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "type", string(imaging.TypeImage)),
					),
				},
			},
		},
		"nok create": {
			init: func(m *mockimaging) {
				contractID, policySetID, policySetName, region, mediaType := "1-TEST", "testID", "test_policy_set", "EMEA", string(imaging.TypeImage)
				createdPolicySet := &imaging.PolicySet{Name: policySetName, ID: policySetID, Region: imaging.Region(region), Type: mediaType}

				// create
				expectPolicySetCreation(t, m, contractID, policySetName, region, mediaType, createdPolicySet, anError)
			},
			steps: []resource.TestStep{
				{
					Config:      loadFixtureString(fmt.Sprintf("%s/lifecycle/create.tf", testDir)),
					ExpectError: regexp.MustCompile(anError.Error()),
				},
			},
		},
		"nok get policy set post create": {
			init: func(m *mockimaging) {
				contractID, policySetID, policySetName, region, mediaType := "1-TEST", "testID", "test_policy_set", "EMEA", string(imaging.TypeImage)
				createdPolicySet := &imaging.PolicySet{Name: policySetName, ID: policySetID, Region: imaging.Region(region), Type: mediaType}

				// create
				expectPolicySetCreation(t, m, contractID, policySetName, region, mediaType, createdPolicySet, nil)

				// create -> read
				expectPolicySetRead(t, m, contractID, policySetID, nil, anError, 1)

				// delete
				expectPolicySetDelete(t, m, contractID, policySetID, "", &imaging.ListPoliciesResponse{
					ItemKind: "POLICY",
					Items: []imaging.PolicyOutput{
						&imaging.PolicyOutputImage{ID: ".auto"},
					},
					TotalItems: 1,
				}, nil, nil)
			},
			steps: []resource.TestStep{
				{
					Config:      loadFixtureString(fmt.Sprintf("%s/lifecycle/create.tf", testDir)),
					ExpectError: regexp.MustCompile(anError.Error()),
				},
			},
		},
		"ok create update": {
			init: func(m *mockimaging) {
				contractID, policySetID, policySetName, EMEA, mediaType, US := "1-TEST", "testID", "test_policy_set", string(imaging.RegionEMEA), string(imaging.TypeImage), string(imaging.RegionUS)
				createdPolicySet := &imaging.PolicySet{Name: policySetName, ID: policySetID, Region: imaging.Region(EMEA), Type: mediaType}
				updatedPolicySet := &imaging.PolicySet{Name: policySetName, ID: policySetID, Region: imaging.Region(US), Type: mediaType}

				// create
				expectPolicySetCreation(t, m, contractID, policySetName, EMEA, mediaType, createdPolicySet, nil)

				// create -> read, test -> read, refresh
				expectPolicySetRead(t, m, contractID, policySetID, createdPolicySet, nil, 3)

				// update
				expectPolicySetUpdate(t, m, contractID, policySetID, policySetName, US, nil)

				// update -> read
				expectPolicySetRead(t, m, contractID, policySetID, updatedPolicySet, nil, 2)

				// delete
				expectPolicySetDelete(t, m, contractID, policySetID, "", &imaging.ListPoliciesResponse{
					ItemKind: "POLICY",
					Items: []imaging.PolicyOutput{
						&imaging.PolicyOutputImage{ID: ".auto"},
					},
					TotalItems: 1,
				}, nil, nil)
			},
			steps: []resource.TestStep{
				{
					Config: loadFixtureString(fmt.Sprintf("%s/lifecycle/create.tf", testDir)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "id", "testID"),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "name", "test_policy_set"),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "region", string(imaging.RegionEMEA)),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "type", string(imaging.TypeImage)),
					),
				},
				{
					Config: loadFixtureString(fmt.Sprintf("%s/lifecycle/update_region_us.tf", testDir)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "id", "testID"),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "name", "test_policy_set"),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "region", string(imaging.RegionUS)),
						resource.TestCheckResourceAttr("akamai_imaging_policy_set.imv_set", "type", string(imaging.TypeImage)),
					),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := &mockimaging{}
			test.init(client)
			useClient(client, func() {
				resource.UnitTest(t, resource.TestCase{
					Providers: testAccProviders,
					Steps:     test.steps,
				})
			})
			client.AssertExpectations(t)
		})
	}
}

func Test_filterRemainingPolicies(t *testing.T) {
	tests := map[string]struct {
		input          *imaging.ListPoliciesResponse
		expectedOutput int
	}{
		"just 1 image policy .auto remaining policy": {
			input: &imaging.ListPoliciesResponse{
				ItemKind: "POLICY",
				Items: []imaging.PolicyOutput{
					&imaging.PolicyOutputImage{ID: ".auto"},
				},
				TotalItems: 0,
			},
			expectedOutput: 0,
		},
		"just 1 video policy .auto remaining policy": {
			input: &imaging.ListPoliciesResponse{
				ItemKind: "POLICY",
				Items: []imaging.PolicyOutput{
					&imaging.PolicyOutputVideo{ID: ".auto"},
				},
				TotalItems: 0,
			},
			expectedOutput: 0,
		},
		"2 video policies, one of them .auto": {
			input: &imaging.ListPoliciesResponse{
				ItemKind: "POLICY",
				Items: []imaging.PolicyOutput{
					&imaging.PolicyOutputVideo{ID: ".auto"},
					&imaging.PolicyOutputVideo{ID: "not-auto"},
				},
				TotalItems: 0,
			},
			expectedOutput: 1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedOutput, filterRemainingPolicies(test.input))
		})
	}

}
