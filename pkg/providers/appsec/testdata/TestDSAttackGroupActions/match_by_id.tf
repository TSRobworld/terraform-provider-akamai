provider "akamai" {
  edgerc = "~/.edgerc"
}


data "akamai_appsec_attack_group_actions" "test" {
config_id = 43253
    version = 7
    security_policy_id = "AAAA_81230"
    attack_group = "SQL"
}



