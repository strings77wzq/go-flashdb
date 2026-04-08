## ADDED Requirements

### Requirement: Master Role Implementation
The system SHALL implement Master node role supporting replication with slave connections.

#### Scenario: Master accepts replication connections
- **WHEN** slave connects to master port
- **THEN** master accepts connection and registers slave in slaves list

#### Scenario: Master tracks replication offset
- **WHEN** master executes write command
- **THEN** master increments replication offset and sends to all connected slaves

#### Scenario: Master handles PSYNC
- **WHEN** slave sends PSYNC with replication offset
- **THEN** master either performs full sync (RDB) or continues with partial sync

### Requirement: Slave Role Implementation
The system SHALL implement Slave node role that synchronizes data from master.

#### Scenario: Slave connects to master
- **WHEN** configured with master_address and starts
- **THEN** slave initiates connection and performs handshake

#### Scenario: Slave receives full sync
- **WHEN** master sends RDB file
- **THEN** slave loads data and transitions to connected state

#### Scenario: Slave receives incremental commands
- **WHEN** master sends replication command stream
- **THEN** slave executes commands in order to maintain consistency

### Requirement: REPLICAOF/SLAVEOF Command
The system SHALL support changing replica configuration via REPLICAOF command.

#### Scenario: Promote to replica
- **WHEN** client sends REPLICAOF host port
- **THEN** node disconnects from current master and connects to new master

#### Scenario: Stop replication
- **WHEN** client sends REPLICAOF NO ONE
- **THEN** node stops being replica and becomes standalone master

### Requirement: ROLE Command
The system SHALL return current node role via ROLE command.

#### Scenario: Query master role
- **WHEN** node is master and client sends ROLE
- **THEN** system returns ["master", offset, [slave1, slave2]]

#### Scenario: Query slave role
- **WHEN** node is slave and client sends ROLE
- **THEN** system returns ["slave", "master_ip", "master_port", "connected"]

### Requirement: Replication Backlog
The system SHALL maintain replication backlog buffer for incremental sync.

#### Scenario: Track replication buffer
- **WHEN** master processes command
- **THEN** command is added to replication backlog with offset

#### Scenario: Incremental sync from offset
- **WHEN** slave requests PSYNC with valid offset
- **THEN** master sends commands from that offset