## ADDED Requirements

### Requirement: Redis Cluster Slot Management
The system SHALL implement Redis Cluster slot management with 16384 slots (0-16383) supporting key distribution across cluster nodes.

#### Scenario: Slot calculation
- **WHEN** client sends command with key "user:1001"
- **THEN** system calculates CRC16(key) mod 16384 to determine slot number

#### Scenario: Slot range query
- **WHEN** client sends CLUSTER SLOTS command
- **THEN** system returns array of [startSlot, endSlot, nodeIP, nodePort] for each node

### Requirement: Cluster Node Information
The system SHALL implement CLUSTER INFO and CLUSTER NODES commands returning cluster state and node topology.

#### Scenario: Cluster info query
- **WHEN** client sends CLUSTER INFO
- **THEN** system returns cluster_state, cluster_slots_assigned, cluster_slots_ok, cluster_known_nodes

#### Scenario: Cluster nodes query
- **WHEN** client sends CLUSTER NODES
- **THEN** system returns detailed node list with role, connected status, slots

### Requirement: MOVED Redirection
The system SHALL implement MOVED redirection when client queries key in different node's slot.

#### Scenario: MOVED response
- **WHEN** client sends GET key and target slot belongs to different node
- **THEN** system returns -MOVED <slot> <nodeAddress> error

#### Scenario: ASK Redirection
- **WHEN** client accesses key during slot migration
- **THEN** system returns -ASK <slot> <nodeAddress> error

### Requirement: Cluster Mode Support
The system SHALL support running in Cluster mode with configurable node role and slot ownership.

#### Scenario: Enable cluster mode
- **WHEN** config has cluster_enabled: true
- **THEN** server starts in cluster mode with CLUSTER commands available

#### Scenario: Configure slot ownership
- **WHEN** admin sends CLUSTER ADDSLOTS command
- **THEN** node claims ownership of specified slot ranges

### Requirement: Client Retry Logic
The system SHALL provide clear error messages enabling client-side retry with proper redirection.

#### Scenario: Handle MOVED error
- **WHEN** client receives MOVED error
- **THEN** client extracts new node address and reconnects to query again

#### Scenario: Handle ASK error
- **WHEN** client receives ASK error
- **THEN** client sends ASKING command to target node before retry