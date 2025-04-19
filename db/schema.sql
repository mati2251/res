CREATE TYPE vm_status AS ENUM ('running', 'stopped', 'paused');

CREATE TABLE vm (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    memory INT NOT NULL DEFAULT 1024,
    cpu INT NOT NULL DEFAULT 1,
    disk INT NOT NULL DEFAULT 20,
    image VARCHAR(255) NOT NULL,
    port INT NOT NULL DEFAULT 2222,
    start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP,
    status vm_status NOT NULL DEFAULT 'running'
);

CREATE TYPE job_status AS ENUM ('success', 'failed', 'running');

CREATE TABLE job (
    id SERIAL PRIMARY KEY,
    vm_id INT NOT NULL,
    status vm_status NOT NULL DEFAULT 'running',
    base_path VARCHAR(255) NOT NULL,
    FOREIGN KEY (vm_id) REFERENCES vm(id)
);


CREATE OR REPLACE FUNCTION update_job_path()
RETURNS TRIGGER AS $$
BEGIN
    NEW.base_path := NEW.base_path || '/' || NEW.id ;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER update_log_path_trigger
BEFORE INSERT ON job
FOR EACH ROW
EXECUTE FUNCTION update_job_path();
