## ADDED Requirements

### Requirement: DBSIZE Command
The system SHALL return the number of keys in current database via DBSIZE command.

#### Scenario: Query database size
- **WHEN** client sends DBSIZE command
- **THEN** system returns integer representing total key count in current database

#### Scenario: Empty database
- **WHEN** database has no keys and client sends DBSIZE
- **THEN** system returns 0

### Requirement: INFO Command
The system SHALL return comprehensive server information and statistics via INFO command.

#### Scenario: INFO server section
- **WHEN** client sends INFO command
- **THEN** system returns server version, uptime, memory usage, connected clients

#### Scenario: INFO memory section
- **WHEN** client sends INFO memory
- **THEN** system returns memory breakdown including used_memory, used_memory_lua

#### Scenario: INFO stats section
- **WHEN** client sends INFO stats
- **THEN** system returns command statistics including total_commands_processed

### Requirement: CONFIG GET Command
The system SHALL support retrieving configuration parameters via CONFIG GET command.

#### Scenario: Get specific config
- **WHEN** client sends CONFIG GET maxclients
- **THEN** system returns array ["maxclients", "10000"]

#### Scenario: Get with wildcard
- **WHEN** client sends CONFIG GET *timeout*
- **THEN** system returns all matching config parameters

### Requirement: CONFIG SET Command
The system SHALL support modifying runtime configuration via CONFIG SET command.

#### Scenario: Set string config
- **WHEN** client sends CONFIG SET maxclients 20000
- **THEN** system returns OK and config is updated

#### Scenario: Set numeric config
- **WHEN** client sends CONFIG SET timeout 300
- **THEN** system returns OK and timeout is changed

### Requirement: FLUSHDB Command
The system SHALL support clearing current database via FLUSHDB command.

#### Scenario: Flush current database
- **WHEN** client sends FLUSHDB
- **THEN** system removes all keys in current database and returns OK

#### Scenario: Flush with async option
- **WHEN** client sends FLUSHDB ASYNC
- **THEN** system returns OK immediately while deletion happens in background

### Requirement: FLUSHALL Command
The system SHALL support clearing all databases via FLUSHALL command.

#### Scenario: Flush all databases
- **WHEN** client sends FLUSHALL
- **THEN** system removes all keys from all databases and returns OK

### Requirement: CLIENT LIST Command
The system SHALL return information about connected clients via CLIENT LIST command.

#### Scenario: List all clients
- **WHEN** client sends CLIENT LIST
- **THEN** system returns formatted list with id, addr, age, flags for each connection