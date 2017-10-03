--
-- Table structure for table `client_location`
--

DROP TABLE IF EXISTS `client_location`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `client_location` (
  `client_location_id` mediumint(8) unsigned NOT NULL AUTO_INCREMENT,
  `client_license_id` mediumint(8) unsigned NOT NULL,
  `client_district_id` mediumint(8) unsigned DEFAULT NULL,
  `title` varchar(100) NOT NULL,
  `descr` varchar(255) NOT NULL,
  `url` varchar(100) DEFAULT NULL,
  `address_line_1` varchar(100) DEFAULT NULL,
  `address_line_2` varchar(100) DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `geo_state_id` smallint(5) unsigned DEFAULT NULL,
  `zip_code` varchar(40) DEFAULT NULL,
  `mdr_location_id` varchar(40) DEFAULT NULL,
  `client_location_id` varchar(36) DEFAULT NULL COMMENT 'Unique location id to give to clients. Use UUID() to insert value.',
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `geo_time_zone_id` varchar(255) DEFAULT 'UTC',
  PRIMARY KEY (`client_location_id`)
) ENGINE=MyISAM AUTO_INCREMENT=1067 DEFAULT CHARSET=utf8 COMMENT='School location';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `client_location`
--

LOCK TABLES `client_location` WRITE;
/*!40000 ALTER TABLE `client_location` DISABLE KEYS */;
INSERT INTO `client_location` VALUES (1,1,1,'AcneSpot School','Test school for AcneSpot','AcneSpotschool','7825 Telegraph Rd','','Bloomington',28,'55438','4875948',NULL,0,'UTC'),(2,1,1,'Edina Middle School','Edina Middle School','edinamiddle','5808 Olinger Blvd',NULL,'Edina',28,'55436','5875948',NULL,0,'UTC'),(3,1,1,'Countryside High School','Countryside High School','countrysidehigh','5701 Benton Avenue',NULL,'Edina',28,'55436-2501','6875948',NULL,0,'UTC'),(325,4,1,'AcneSpot Middle School','AcneSpot Middle School','AcneSpotmiddle','7825 Telegraph Road','','Bloomington',28,'55438','1000001',NULL,0,'UTC');
/*!40000 ALTER TABLE `client_location` ENABLE KEYS */;
UNLOCK TABLES;
