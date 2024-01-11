/*-------------------------------------------------------------------------
 *
 * pg_authid_d.h
 *    Macro definitions for pg_authid
 *
 * Portions Copyright (c) 1996-2023, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 * NOTES
 *  ******************************
 *  *** DO NOT EDIT THIS FILE! ***
 *  ******************************
 *
 *  It has been GENERATED by src/backend/catalog/genbki.pl
 *
 *-------------------------------------------------------------------------
 */
#ifndef PG_AUTHID_D_H
#define PG_AUTHID_D_H

#define AuthIdRelationId 1260
#define AuthIdRelation_Rowtype_Id 2842
#define PgAuthidToastTable 4175
#define PgAuthidToastIndex 4176
#define AuthIdRolnameIndexId 2676
#define AuthIdOidIndexId 2677

#define Anum_pg_authid_oid 1
#define Anum_pg_authid_rolname 2
#define Anum_pg_authid_rolsuper 3
#define Anum_pg_authid_rolinherit 4
#define Anum_pg_authid_rolcreaterole 5
#define Anum_pg_authid_rolcreatedb 6
#define Anum_pg_authid_rolcanlogin 7
#define Anum_pg_authid_rolreplication 8
#define Anum_pg_authid_rolbypassrls 9
#define Anum_pg_authid_rolconnlimit 10
#define Anum_pg_authid_rolpassword 11
#define Anum_pg_authid_rolvaliduntil 12

#define Natts_pg_authid 12

#define BOOTSTRAP_SUPERUSERID 10
#define ROLE_PG_DATABASE_OWNER 6171
#define ROLE_PG_READ_ALL_DATA 6181
#define ROLE_PG_WRITE_ALL_DATA 6182
#define ROLE_PG_MONITOR 3373
#define ROLE_PG_READ_ALL_SETTINGS 3374
#define ROLE_PG_READ_ALL_STATS 3375
#define ROLE_PG_STAT_SCAN_TABLES 3377
#define ROLE_PG_READ_SERVER_FILES 4569
#define ROLE_PG_WRITE_SERVER_FILES 4570
#define ROLE_PG_EXECUTE_SERVER_PROGRAM 4571
#define ROLE_PG_SIGNAL_BACKEND 4200
#define ROLE_PG_CHECKPOINT 4544
#define ROLE_PG_USE_RESERVED_CONNECTIONS 4550
#define ROLE_PG_CREATE_SUBSCRIPTION 6304

#endif							/* PG_AUTHID_D_H */