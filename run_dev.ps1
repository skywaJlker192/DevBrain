# run_dev.ps1
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="postgres"
$env:DB_NAME="devbrain"
$env:DB_SSLMODE="disable"
$env:SERVER_PORT="8080"
$env:ENV="development"
$env:ALLOWED_ORIGINS="http://localhost:8080"
$env:MAX_BODY_SIZE="10485760"
$env:JWT_SECRET="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InRlc3RAZGV2YnJhaW4ucnUiLCJyb2xlIjoidXNlciIsImlzcyI6ImRldmJyYWluLXBybyIsInN1YiI6InVzZXI6MSIsImV4cCI6MTc4MTAyOTQ5OSwiaWF0IjoxNzgxMDI3Njk5fQ.Y8r0Al1WZWwB1LnodHvHzYWlVLRRIb1W7stLnvcU-qE"
$env:JWT_ISSUER="devbrain-pro"
$env:JWT_ACCESS_TTL="30m"
$env:JWT_REFRESH_TTL="168h"
$env:RATE_LIMIT=10
$env:RATE_LIMIT_WINDOW="1m"
$env:LOG_LEVEL="info"

& "C:\Program Files\Go\bin\go.exe" run ./cmd/server
