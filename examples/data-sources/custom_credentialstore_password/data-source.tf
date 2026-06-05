# read a simple credential password
data "custom_credentialstore_password" "main" {
  namespace = "some-namespace"
  name      = "some-password"
}