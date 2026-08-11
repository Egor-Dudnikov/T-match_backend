COPY cities (name, region, geo_lat, geo_lon)
FROM '/docker-entrypoint-initdb.d/data/cities.csv'
DELIMITER ','
CSV HEADER;

COPY skills (name)
FROM '/docker-entrypoint-initdb.d/data/skills.csv'
DELIMITER ','
CSV HEADER;