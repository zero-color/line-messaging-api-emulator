env "local" {
  // Declare where the schema definition resides.
  // Also supported: ["file://multi.hcl", "file://schema.hcl"].
  src = "file://schema.sql"

  // Define the URL of the database which is managed
  // in this environment.
  url = "postgres://postgres:password@127.0.0.1:5432/line_emulator?search_path=public&sslmode=disable"
  
  // Define the dev database URL for schema development
  dev = "docker://postgres/17/dev?search_path=public"
}
