## ADDED Requirements

### Requirement: WATCH Command
The system SHALL implement optimistic locking via WATCH command.

#### Scenario: Watch single key
- **WHEN** client sends WATCH key
- **THEN** key is added to watch list for current connection

#### Scenario: Watch multiple keys
- **WHEN** client sends WATCH key1 key2 key3
- **THEN** all keys are added to watch list

#### Scenario: Watch detects modification
- **WHEN** watched key is modified before EXEC
- **THEN** transaction execution returns null (empty array)

### Requirement: MULTI Command
The system SHALL start transaction mode via MULTI command.

#### Scenario: Start transaction
- **WHEN** client sends MULTI
- **THEN** system enters transaction mode and returns OK

#### Scenario: Queue commands in transaction
- **WHEN** client sends SET in transaction mode
- **THEN** command is queued and returns QUEUED

### Requirement: EXEC Command
The system SHALL execute transaction and return all command results.

#### Scenario: Execute successful transaction
- **WHEN** client sends EXEC after queued commands
- **THEN** system returns array of each command's result

#### Scenario: Execute failed transaction
- **WHEN** transaction contains error and client sends EXEC
- **THEN** system returns array with error for that command position

### Requirement: DISCARD Command
The system SHALL cancel transaction and clear command queue.

#### Scenario: Discard transaction
- **WHEN** client sends DISCARD in transaction mode
- **THEN** all queued commands are discarded and transaction ends

### Requirement: Transaction Atomicity
The system SHALL guarantee atomic execution of all commands in transaction.

#### Scenario: Transaction atomic execution
- **WHEN** EXEC is called
- **THEN** either all commands succeed or none are applied

#### Scenario: Watch key changed
- **WHEN** watched key modified and EXEC called
- **THEN** transaction is cancelled, all commands not executed