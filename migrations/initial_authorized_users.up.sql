BEGIN TRANSACTION;
INSERT INTO users (id, email, role) VALUES ('70cf16d3-de03-4726-8e48-02cc6c93afa5'
                                           ,'hristo_partenov@abv.bg','admin');
INSERT INTO users (id,email,role) VALUES ('d8ee1ae0-4415-453d-87c4-60dedf60f67e'
                                         ,'ivan@yahoo.com','writer');
INSERT INTO users (id,email,role) VALUES ('4bce0bff-0cd5-4cf6-b64c-847d45798034'
                                         ,'eva@gmail.com','reader');
COMMIT;