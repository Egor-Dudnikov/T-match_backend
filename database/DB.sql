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