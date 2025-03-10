# Calenduh API

## Team Information
**Team Number:** 5  
**Team Members:**
- Kaden Hardesty
- Edward Navarro
- Evan Borszem
- Matthew Gardner
- Kellen Neff

## Table of Contents
- [System Requirements](#system-requirements)
- [Installation](#installation)
- [Project Setup](#project-setup)
- [Environment Configuration](#environment-configuration)
- [Development Tools](#development-tools)
- [Build and Run Instructions](#build-and-run-instructions)
- [Debugging and Troubleshooting](#debugging-and-troubleshooting)
- [Deployment](#deployment)

## System Requirements

### Hardware Specifications
- **CPU:** 2+ cores recommended (1 core minimum)
- **RAM:** 4GB minimum, 8GB+ recommended
- **Storage:** 5GB free disk space minimum, 10GB+ recommended

### Software Dependencies
- **Operating System:**
   - macOS 10.15+
   - Windows 10+ with WSL2
- **Core Software:**
   - [Docker Desktop](https://docs.docker.com/desktop/release-notes) (Latest)
   - [Go](https://go.dev/dl/) (version 1.22.9)
- **Required CLI Tools:**
   - make (Requires Git Bash and Chocolatey on Windows)
   - [Golang Migrate](https://github.com/golang-migrate/migrate) (Latest)
   - [SQLC](https://docs.sqlc.dev/en/stable/overview/install.html) (Latest)

## Installation of Requisite Software

### Git Bash + Make (Windows Only)

1. [Download](https://git-scm.com/downloads/win) and run the Git Bash installer 
2. Inspect the [installation script](https://community.chocolatey.org/install.ps1) for Chocolatey prior to execution for safety
3. Open an instance of Powershell **as Administrator** and execute the installation script
   ```ps
   Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
   ```
4. In a new instance of Powershell install make with chocolatey
   ```ps
   choco install make
   ```

### Docker Desktop

1. [Download](https://docs.docker.com/desktop/release-notes) and run the Docker Desktop installer

### Golang

1. [Download](https://go.dev/dl/) and run the Golang installer for the desired version

### Golang Migrate

1. Follow the [installation instructions](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md)
for golang migrate using any of the listed options for your operating system

### SQLC

1. Follow the default [installation instructions](https://docs.sqlc.dev/en/stable/overview/install.html)
   for sqlc using any of the listed options for your operating system

## Project Setup & Configuration

### \*It is recommended to use a modern editor like Intellij or VSCode from this point on

1. [Clone or download](https://github.com/TechnoWizzy/calenduh-backend) the repository locally
2. Configure your editor to support Golang type and syntax highlighting, using the Golang SDK you installed previously. This may require the installation of 3rd-party
plugins.
3. Create a file called `.env` in the project root and populate these default values
   ```
    POSTGRESQL_URL=postgresql://username:password@calenduh-db:5432/calenduh?sslmode=disable
    POSTGRES_USER=username
    POSTGRES_PASSWORD=password
    POSTGRES_DB=calenduh
    ```
4. Use SQLC to generate database interface files in `/internal/sqlc/`. This generated code will already be referenced throughout the project.
   ```bash
   make sqlc
   ```

## Building & Running

### \*It is required that port 8080 and 8962 not be in-use by any other application. These values can be modified in ./docker-compose.yml, and will be important for reaching the application and database for testing.

1. Build the containers
   ```bash
   make build
   ```
2. Run the containers
   ```bash
   make run
   ```
3. Delete the containers
   ```bash
   make clean
   ```
4. Delete the containers and persistent volumes (useful if database is in an errored state)
   ```bash
   make clean-db
   ```
5. Creating a new database migration
   ```bash
   make migration
   ```
6. Regenerate the database types in Golang
   ```bash
   make sqlc
   ```