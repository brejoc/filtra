-- PostgreSQL (denormalized) schema

-- Create user and database (example):
--
-- create role filtra with password 'filtra';
-- alter role filtra with superuser login;
-- create database filtra;
-- grant all privileges on database filtra to filtra;


-- Types:
-- * ALL
-- * BLOCKED
-- * CLOSED
-- * IN_PROGRESS
-- * OPEN_ISSUE
-- * OPEN_BUG
-- * OPEN_L3_BUG
-- * PLANNED

CREATE TABLE issue_counter(
	id serial PRIMARY KEY,
	ts timestamp(4) with time zone NOT NULL,
	type varchar(255) NOT NULL,
	value int NOT NULL
);

-- Types:
-- * CYCLE_TIME
-- * LEAD_TIME

CREATE TABLE issue_flow(
	id serial PRIMARY KEY,
	ts timestamp(4) with time zone NOT NULL,
	type varchar(255) NOT NULL,
	value float NOT NULL
);
