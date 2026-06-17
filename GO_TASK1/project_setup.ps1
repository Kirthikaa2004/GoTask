# PowerShell script for project setup (Windows)

$SCRIPT_DIR = $PSScriptRoot
$PROJECT_ROOT = $SCRIPT_DIR
$SERVICES_DIR = Join-Path $PROJECT_ROOT "services"

function EnsureDir($path) {
    if (!(Test-Path $path)) { New-Item -ItemType Directory -Force -Path $path | Out-Null }
}

function EnsureFile($path) {
    EnsureDir (Split-Path $path -Parent)
    if (!(Test-Path $path)) { New-Item -ItemType File -Force -Path $path | Out-Null }
}

function CreateService {
    $project_name = Read-Host "Enter the root project name"
    $service_name = Read-Host "Enter the name of the new service"
    $entity = Read-Host "Enter the entity name (e.g. user, book, order)"
    $svc = Join-Path $SERVICES_DIR $service_name

    # cmd
    EnsureFile "$svc\cmd\server\main.go"

    # deployments
    EnsureFile "$svc\deployments\Dockerfile"
    EnsureFile "$svc\deployments\docker-compose.yml"

    # docs
    EnsureFile "$svc\docs\openapi.yaml"
    EnsureFile "$svc\docs\swaggerui\index.html"
    EnsureDir "$svc\docs\swaggerui\scripts"

    # envs
    EnsureFile "$svc\envs\.env.dev"

    # internal
    EnsureFile "$svc\internal\config\config.go"
    EnsureFile "$svc\internal\dtos\response_dto.go"
    EnsureFile "$svc\internal\dtos\${entity}_dto.go"
    EnsureFile "$svc\internal\errorcodes\error_codes.go"
    EnsureFile "$svc\internal\handlers\routes.go"
    EnsureFile "$svc\internal\handlers\${entity}_handler.go"
    EnsureFile "$svc\internal\loggers\logger.go"
    EnsureFile "$svc\internal\models\${entity}.go"
    EnsureFile "$svc\internal\repository\${entity}_repository.go"
    EnsureFile "$svc\internal\services\${entity}_service.go"
    EnsureFile "$svc\internal\utils\utils.go"

    # pkg
    EnsureFile "$svc\pkg\database\database.go"
    EnsureFile "$svc\pkg\database\migration.go"
    EnsureFile "$svc\pkg\server\server.go"

    # tests
    EnsureFile "$svc\tests\integration\.gitkeep"
    EnsureFile "$svc\tests\unit\config\config_test.go"
    EnsureFile "$svc\tests\unit\config\mocks\config_mock.go"
    EnsureFile "$svc\tests\unit\database\database_test.go"
    EnsureFile "$svc\tests\unit\handlers\${entity}_handler_test.go"
    EnsureFile "$svc\tests\unit\handlers\routes_test.go"
    EnsureFile "$svc\tests\unit\pkg\server\server_test.go"
    EnsureFile "$svc\tests\unit\repository\${entity}_repository_test.go"
    EnsureFile "$svc\tests\unit\repository\mock\${entity}_repository_mock.go"
    EnsureFile "$svc\tests\unit\services\${entity}_service_test.go"
    EnsureFile "$svc\tests\unit\services\mock\${entity}_service_mock.go"
    EnsureFile "$svc\tests\unit\server\server_test.go"
    EnsureFile "$svc\tests\unit\utils\utils_test.go"

    # go module
    Set-Location $svc
    go mod init "$project_name/$service_name"
    go get github.com/gorilla/mux
    go get go.uber.org/zap
    go get github.com/joho/godotenv
    go get github.com/stretchr/testify/assert
    go get github.com/lib/pq
    go mod tidy
    Set-Location $PROJECT_ROOT

    EnsureFile "$svc\README.md"
    Write-Host "Service '$service_name' created at $svc"
}

function ShowMenu {
    Write-Host "1. Create service"
    Write-Host "2. Exit"
    $choice = Read-Host "Enter your choice"
    switch ($choice) {
        "1" { CreateService }
        "2" { return }
        default { Write-Host "Invalid choice." }
    }
    ShowMenu
}

ShowMenu
