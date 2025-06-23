-- Initialize databases for impl-zamaz Zero Trust demo
-- This script runs automatically when PostgreSQL container starts

-- Create Keycloak database and user
CREATE DATABASE keycloak;
CREATE USER keycloak WITH ENCRYPTED PASSWORD 'keycloak_password';
GRANT ALL PRIVILEGES ON DATABASE keycloak TO keycloak;

-- Create application database (if needed)
CREATE DATABASE impl_zamaz;
GRANT ALL PRIVILEGES ON DATABASE impl_zamaz TO postgres;

-- Log initialization
SELECT 'Database initialization completed' AS status;