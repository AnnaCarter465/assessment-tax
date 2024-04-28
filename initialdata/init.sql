SELECT 'CREATE DATABASE ktaxes'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ktaxes')\gexec

\c "ktaxes";

CREATE TABLE IF NOT EXISTS default_allowances (
    allowance_type varchar(100) NOT NULL,
    amount float8 DEFAULT 0 NOT NULL,
    CONSTRAINT default_allowances_pk PRIMARY KEY (allowance_type)
);	

CREATE TABLE IF NOT EXISTS allowed_allowances (
    allowance_type varchar(100) NOT NULL,
    max_amount float8 DEFAULT 0 NOT NULL,
    CONSTRAINT allowed_allowances_pk PRIMARY KEY (allowance_type)
);	


INSERT INTO default_allowances (allowance_type,amount)
VALUES ('personal',60000.0)
ON CONFLICT (allowance_type) DO NOTHING;	


INSERT INTO allowed_allowances (allowance_type,max_amount)
VALUES 
    ('donation',100000.0),
    ('k-receipt',50000.0)
ON CONFLICT (allowance_type) DO NOTHING;