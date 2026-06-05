# create a simple credential password
resource "custom_credentialstore_password" "main" {
  namespace = "some-namespace"
  name      = "some-password"
  value     = "my-s3cr3t"
}