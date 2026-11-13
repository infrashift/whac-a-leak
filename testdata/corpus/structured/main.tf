# Every value is fake.
resource "aws_db_instance" "main" {
  username = "app"
  password = "EXAMPLE-hunter2"
}

resource "aws_db_instance" "replica" {
  password = var.replica_password
}
