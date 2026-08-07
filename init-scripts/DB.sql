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

CREATE TYPE "notification_type" AS ENUM (
  'invate',
  'change_status',
  'new_application'
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
  "user_id" INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
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
  "user_id" INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  "company_name" VARCHAR UNIQUE NOT NULL,
  "description" TEXT,
  "website" VARCHAR,
  "inn" VARCHAR(12) UNIQUE,
  "kpp" VARCHAR(9),
  "ogrn" VARCHAR(15),
  "legal_address" TEXT,
  "director_name" VARCHAR,
  "image" TEXT
);

CREATE TABLE "admins" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  "name" VARCHAR
);

CREATE TABLE "internships" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "company_id" INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  "title" VARCHAR,
  "description" TEXT,
  "salary" INTEGER,
  "duration_months" INTEGER,
  "location" VARCHAR,
  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  "is_archived" BOOLEAN NOT NULL DEFAULT FALSE,
  "tsv_content" tsvector
);

CREATE TABLE "applications" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "intern_id" INTEGER NOT NULL REFERENCES interns(id) ON DELETE CASCADE,
  "internship_id" INTEGER NOT NULL REFERENCES internships(id) ON DELETE CASCADE,
  "status" APPLICATION_STATUS DEFAULT 'pending',
  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "skills" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "name" VARCHAR UNIQUE
);

CREATE TABLE "intern_skills" (
  "intern_id" INTEGER NOT NULL REFERENCES interns(id) ON DELETE CASCADE,
  "skill_id" INTEGER NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  PRIMARY KEY ("intern_id", "skill_id")
);

CREATE TABLE "internship_skills" (
  "internship_id" INTEGER NOT NULL REFERENCES internships(id) ON DELETE CASCADE,
  "skill_id" INTEGER NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  PRIMARY KEY ("internship_id", "skill_id")
);

CREATE TABLE "notifications" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  "type" NOTIFICATION_TYPE,
  "is_read" BOOLEAN NOT NULL DEFAULT FALSE,
  "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "change_status_data" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "notification_id" INTEGER NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
  "internship_id" INTEGER NOT NULL REFERENCES internships(id) ON DELETE CASCADE,
  "company_id" INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  "new_status" VARCHAR NOT NULL
);

CREATE TABLE "invate_data" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "notification_id" INTEGER NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
  "internship_id" INTEGER NOT NULL REFERENCES internships(id) ON DELETE CASCADE,
  "company_id" INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  "message" VARCHAR
);

CREATE TABLE "new_application_data" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "notification_id" INTEGER NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
  "internship_id" INTEGER NOT NULL REFERENCES internships(id) ON DELETE CASCADE,
  "intern_id" INTEGER NOT NULL REFERENCES interns(id) ON DELETE CASCADE,
  "response_id" INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE
);

CREATE TABLE "user_bans" (
    "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id" INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    "reason" TEXT NOT NULL,
    "banned_by" INTEGER NOT NULL REFERENCES users(id),
    "banned_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ON "applications" ("intern_id", "internship_id");

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

CREATE INDEX idx_applications_intern_id ON applications(intern_id);
CREATE INDEX idx_applications_internship_id ON applications(internship_id);
CREATE INDEX idx_applications_status ON applications(status);

CREATE INDEX idx_internships_company_id ON internships(company_id);
CREATE INDEX idx_internships_created_at ON internships(created_at DESC);
CREATE INDEX idx_internships_is_archived ON internships(is_archived);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

CREATE INDEX idx_user_bans_user_id ON user_bans(user_id);

CREATE INDEX idx_intern_skills_skill_id ON intern_skills(skill_id);
CREATE INDEX idx_internship_skills_skill_id ON internship_skills(skill_id);

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
    setweight(to_tsvector('russian', coalesce(NEW.company_name, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B') ||
    setweight(to_tsvector('russian', coalesce(NEW.legal_address, '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER companies_tsv_trigger
BEFORE INSERT OR UPDATE ON companies
FOR EACH ROW EXECUTE FUNCTION update_company_search_vector();

ALTER TABLE interns
ADD COLUMN tsv_content tsvector;

CREATE OR REPLACE FUNCTION update_intern_search_vector() RETURNS trigger AS $$
BEGIN
  NEW.tsv_content :=
    setweight(to_tsvector('russian', coalesce(NEW.first_name, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(NEW.last_name, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(NEW.bio, '')), 'B') ||
    setweight(to_tsvector('russian', coalesce(NEW.experience, '')), 'C') ||
    setweight(to_tsvector('russian', coalesce(NEW.degree, '')), 'C') ||
    setweight(to_tsvector('russian', coalesce(NEW.university , '')), 'D');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER intern_tsv_trigger
BEFORE INSERT OR UPDATE ON interns
FOR EACH ROW EXECUTE FUNCTION update_intern_search_vector();

CREATE INDEX idx_intenship_tsv ON internships USING GIN(tsv_content);
CREATE INDEX idx_companies_tsv ON companies USING GIN(tsv_content);
CREATE INDEX idx_interns_tsv ON interns USING GIN(tsv_content);
