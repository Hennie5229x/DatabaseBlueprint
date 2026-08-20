# SQL Server database objects

This list is limited to objects that belong to a database schema. Server setup,
logins, users, permissions, linked servers, and other infrastructure are outside
Blueprint's scope.

## Recommended creation order

1. Schemas
2. Alias user-defined data types
3. User-defined table types
4. Sequences
5. Synonyms
6. Functions required by table definitions
7. Tables, including columns, defaults, checks, primary keys, unique constraints, and indexes
8. Foreign keys
9. Remaining scalar and table-valued functions
10. Views
11. Stored procedures
12. DML triggers
13. Table data (handled separately from schema `RunOrder.json`)

The dependency graph should refine this order where necessary. For example, a
function used by a table default must be created before that table.

## Export coverage

- [ ] Schemas
- [x] Alias user-defined data types
- [x] User-defined table types
- [ ] Sequences
- [x] Synonyms
- [x] Tables
  - [x] Columns and computed columns
  - [x] Default constraints
  - [x] Check constraints
  - [x] Primary keys
  - [x] Unique constraints
  - [x] Indexes
- [x] Foreign keys
- [x] Scalar functions
- [x] Inline table-valued functions
- [x] Multi-statement table-valued functions
- [x] Views
- [x] Stored procedures
- [ ] DML triggers
- [x] Table data

## Explicitly out of scope for now

- Server-level logins and configuration
- Database users, roles, and permissions
- Linked servers and external server configuration
- CLR assemblies, CLR types, CLR functions, and CLR aggregates
- XML schema collections
- Legacy standalone defaults and rules
- Database-level DDL/event triggers
