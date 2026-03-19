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
  "email" varchar UNIQUE NOT NULL,
  "password_hash" varchar,
  "role" user_role,
  "created_at" timestamp
);

CREATE TABLE "interns" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" integer NOT NULL,
  "first_name" varchar,
  "last_name" varchar,
  "birth_date" date,
  "location" varchar,
  "university" varchar,
  "degree" varchar,
  "bio" text,
  "experience" text
);

CREATE TABLE "companies" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" integer NOT NULL,
  "company_name" varchar UNIQUE NOT NULL,
  "description" text,
  "website" varchar
);

CREATE TABLE "admins" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" integer NOT NULL,
  "name" varchar
);

CREATE TABLE "internships" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "company_id" integer NOT NULL,
  "title" varchar,
  "description" text,
  "salary" integer,
  "duration_months" integer,
  "location" varchar,
  "created_at" timestamp
);

CREATE TABLE "applications" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "intern_id" integer NOT NULL,
  "internship_id" integer NOT NULL,
  "status" application_status DEFAULT 'pending',
  "created_at" timestamp
);

CREATE TABLE "skills" (
  "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "name" varchar UNIQUE
);

CREATE TABLE "intern_skills" (
  "intern_id" integer NOT NULL,
  "skill_id" integer NOT NULL,
  PRIMARY KEY ("intern_id", "skill_id")
);

CREATE TABLE "internship_skills" (
  "internship_id" integer NOT NULL,
  "skill_id" integer NOT NULL,
  PRIMARY KEY ("internship_id", "skill_id")
);

CREATE UNIQUE INDEX ON "applications" ("intern_id", "internship_id");

ALTER TABLE "companies" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "interns" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "admins" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "internships" ADD FOREIGN KEY ("company_id") REFERENCES "companies" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "applications" ADD FOREIGN KEY ("intern_id") REFERENCES "interns" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "applications" ADD FOREIGN KEY ("internship_id") REFERENCES "internships" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "internship_skills" ADD FOREIGN KEY ("skill_id") REFERENCES "skills" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "internship_skills" ADD FOREIGN KEY ("internship_id") REFERENCES "internships" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "intern_skills" ADD FOREIGN KEY ("intern_id") REFERENCES "interns" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "intern_skills" ADD FOREIGN KEY ("skill_id") REFERENCES "skills" ("id") DEFERRABLE INITIALLY IMMEDIATE;