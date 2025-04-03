CREATE TYPE vm_status AS ENUM ('running', 'stopped', 'paused');

CREATE TABLE vm (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    memory INT,
    cpu INT,
    disk INT,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    status vm_status
);

CREATE TYPE job_status AS ENUM ('success', 'failed', 'running');

CREATE TABLE job (
    id SERIAL PRIMARY KEY,
    vm_id INT,
    status vm_status,
    script PATH,
    log PATH,
    FOREIGN KEY (vm_id) REFERENCES vm(id)
);
