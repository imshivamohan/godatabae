CREATE TABLE technical_identities (
    identity TEXT PRIMARY KEY
);

CREATE TABLE data_domains (
    domain_name TEXT PRIMARY KEY
);

CREATE TABLE data_domain_identities (
    identity TEXT REFERENCES technical_identities(identity),
    domain_name TEXT REFERENCES data_domains(domain_name),
    PRIMARY KEY (identity, domain_name)
);
