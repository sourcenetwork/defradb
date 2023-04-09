####################################################################################
# TAG CONFIGURATION SETTINGS
####################################################################################
organization = "source"                     # organizational identifier tag 
environment = "prod"                         # deployment environment tag
deployment_type = "terraform"             # deployment type [ terrform ] 
deployment_repo = "" # location of template repo

####################################################################################
# EC@ CONFIGURATION SETTINGS
####################################################################################

# ec2 region
region = "us-east-1"                      # deployment region 

#key
keypair = "ci-test-triumph"

# security group
ports = [
    {
      from   = 1981
      to     = 1981
      source = "0.0.0.0/0"
    }
  ]
