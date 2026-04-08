## ADDED Requirements

### Requirement: Git Commit and Push
All code changes SHALL be committed and pushed to remote repository.

#### Scenario: Commit all changes
- **WHEN** feature implementation is complete
- **THEN** all new and modified files are committed with descriptive message

#### Scenario: Push to remote
- **WHEN** commit is created
- **THEN** changes are pushed to GitHub remote repository

### Requirement: Version Tagging
Each release SHALL have a semantic version tag.

#### Scenario: Create version tag
- **WHEN** release is ready
- **THEN** git tag with version number (e.g., v0.3.0) is created

#### Scenario: Push tag to remote
- **WHEN** tag is created
- **THEN** tag is pushed to GitHub

### Requirement: GitHub Release
Each version SHALL have a GitHub Release with changelog.

#### Scenario: Create GitHub Release
- **WHEN** version tag is pushed
- **THEN** GitHub Release is created with changelog

#### Scenario: Release includes binaries
- **WHEN** creating release
- **THEN** release includes compiled binaries for major platforms