CREATE TYPE "application_status" AS ENUM (
  'pending',
  'reviewing',
  'accepted',
  'rejected'
);

CREATE TYPE "user_role" AS ENUM (
  'intern',
  'company',
  'admin'
);

CREATE TABLE "users" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "email" VARCHAR UNIQUE NOT NULL,
  "password_hash" VARCHAR,
  "role" USER_ROLE,
  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "interns" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" INTEGER NOT NULL,
  "first_name" VARCHAR,
  "last_name" VARCHAR,
  "birth_date" DATE,
  "location" VARCHAR,
  "university" VARCHAR,
  "degree" VARCHAR,
  "bio" TEXT,
  "experience" TEXT,
  "image" TEXT  
);

CREATE TABLE "companies" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" INTEGER NOT NULL,
  "company_name" VARCHAR UNIQUE NOT NULL,
  "description" TEXT,
  "website" VARCHAR,
  "inn" VARCHAR(12) UNIQUE,
  "kpp" VARCHAR(9),
  "ogrn" VARCHAR(15),
  "legal_address" TEXT,
  "director_name" VARCHAR
);

CREATE TABLE "admins" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" INTEGER NOT NULL,
  "name" VARCHAR
);

CREATE TABLE "internships" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "company_id" INTEGER NOT NULL,
  "title" VARCHAR,
  "description" TEXT,
  "salary" INTEGER,
  "duration_months" INTEGER,
  "location" VARCHAR,
  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "applications" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "intern_id" INTEGER NOT NULL,
  "internship_id" INTEGER NOT NULL,
  "status" APPLICATION_STATUS DEFAULT 'pending',
  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "skills" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "name" VARCHAR UNIQUE
);

CREATE TABLE "intern_skills" (
  "intern_id" INTEGER NOT NULL,
  "skill_id" INTEGER NOT NULL,
  PRIMARY KEY ("intern_id", "skill_id")
);

CREATE TABLE "internship_skills" (
  "internship_id" INTEGER NOT NULL,
  "skill_id" INTEGER NOT NULL,
  PRIMARY KEY ("internship_id", "skill_id")
);

CREATE UNIQUE INDEX ON "applications" ("intern_id", "internship_id");

-- Foreign keys
ALTER TABLE "interns" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "companies" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "admins" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "internships" ADD FOREIGN KEY ("company_id") REFERENCES "companies" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "applications" ADD FOREIGN KEY ("intern_id") REFERENCES "interns" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "applications" ADD FOREIGN KEY ("internship_id") REFERENCES "internships" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "intern_skills" ADD FOREIGN KEY ("intern_id") REFERENCES "interns" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "intern_skills" ADD FOREIGN KEY ("skill_id") REFERENCES "skills" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "internship_skills" ADD FOREIGN KEY ("internship_id") REFERENCES "internships" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "internship_skills" ADD FOREIGN KEY ("skill_id") REFERENCES "skills" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE companies ADD COLUMN IF NOT EXISTS image TEXT;

ALTER TABLE internships 
ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE internships ADD COLUMN tsv_content tsvector;
    
CREATE OR REPLACE FUNCTION update_internships_search_vector() RETURNS trigger AS $$
BEGIN
  NEW.tsv_content := 
  setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') || 
  setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B') ||
  setweight(to_tsvector('russian', coalesce(NEW.location, '')), 'C') ||
  setweight(to_tsvector('russian', coalesce((SELECT company_name FROM companies WHERE id = NEW.company_id), '')), 'D');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER internship_tsv_trigger
BEFORE INSERT OR UPDATE ON internships
FOR EACH ROW EXECUTE FUNCTION update_internships_search_vector();

ALTER TABLE companies
ADD COLUMN tsv_content tsvector;

CREATE OR REPLACE FUNCTION update_company_search_vector() RETURNS trigger AS $$
BEGIN
  NEW.tsv_content := 
    setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B') ||
    setweight(to_tsvector('russian', coalesce(NEW.legal_address, '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER companies_tsv_trigger
BEFORE INSERT OR UPDATE ON companies
FOR EACH ROW EXECUTE FUNCTION update_company_search_vector();
