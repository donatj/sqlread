CREATE TABLE `records` (
  `id` int NOT NULL,
  `name` varchar(255) NOT NULL
);
INSERT INTO `records` VALUES (1,'alpha'),(2,'two,parts');
CREATE TABLE `other` (
  `id` int NOT NULL
);
INSERT INTO `other` VALUES (99);
INSERT INTO `records` VALUES (3,'three\nlines'),(4,'four');
INSERT INTO `records` VALUES (5,'five'),(6,'six');
