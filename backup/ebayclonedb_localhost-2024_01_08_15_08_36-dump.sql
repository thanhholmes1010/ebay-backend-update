-- MySQL dump 10.13  Distrib 8.0.35, for Linux (x86_64)
--
-- Host: 127.0.0.1    Database: ebayclonedb
-- ------------------------------------------------------
-- Server version	8.0.35

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `products`
--

DROP TABLE IF EXISTS `products`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `products` (
  `ProductTypeId` int unsigned NOT NULL,
  `Id` int unsigned NOT NULL AUTO_INCREMENT,
  `Name` varchar(40) NOT NULL,
  `AggregateFields` json NOT NULL,
  PRIMARY KEY (`Id`),
  KEY `product_type_fk` (`ProductTypeId`),
  CONSTRAINT `product_type_fk` FOREIGN KEY (`ProductTypeId`) REFERENCES `producttypes` (`Id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `products`
--

LOCK TABLES `products` WRITE;
/*!40000 ALTER TABLE `products` DISABLE KEYS */;
INSERT INTO `products` VALUES (5,2,'thaianh','{\"fields\": \"hello\"}');
/*!40000 ALTER TABLE `products` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `producttypes`
--

DROP TABLE IF EXISTS `producttypes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `producttypes` (
  `Id` int unsigned NOT NULL AUTO_INCREMENT,
  `Name` varchar(40) NOT NULL,
  `Attributes` json NOT NULL,
  PRIMARY KEY (`Id`),
  UNIQUE KEY `table_name_pk2` (`Name`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `producttypes`
--

LOCK TABLES `producttypes` WRITE;
/*!40000 ALTER TABLE `producttypes` DISABLE KEYS */;
INSERT INTO `producttypes` VALUES (5,'xe-container','{\"attributes\": [{\"id\": 1, \"name\": \"trong-tai\", \"option_values\": [{\"id\": 1, \"value\": \"10T\"}, {\"id\": 2, \"value\": \"20T\"}, {\"id\": 3, \"value\": \"30T\"}]}, {\"id\": 2, \"name\": \"color\", \"option_values\": [{\"id\": 1, \"value\": \"blue\"}, {\"id\": 2, \"value\": \"grey\"}, {\"id\": 3, \"value\": \"red\"}]}]}'),(6,'laptop','{\"attributes\": [{\"id\": 1, \"name\": \"ram\", \"option_values\": [{\"id\": 1, \"value\": \"8G\"}, {\"id\": 2, \"value\": \"16GB\"}]}, {\"id\": 2, \"name\": \"color\", \"option_values\": [{\"id\": 1, \"value\": \"grey\"}, {\"id\": 2, \"value\": \"pink\"}, {\"id\": 3, \"value\": \"black\"}]}, {\"id\": 3, \"name\": \"brand\", \"option_values\": [{\"id\": 1, \"value\": \"apple\"}, {\"id\": 2, \"value\": \"dell\"}, {\"id\": 3, \"value\": \"asus\"}]}, {\"id\": 4, \"name\": \"battery\", \"option_values\": [{\"id\": 1, \"value\": \"over 5 hour\"}, {\"id\": 2, \"value\": \"over 7 hour\"}]}]}'),(7,'xedap','{\"attributes\": [{\"id\": 1, \"name\": \"kilogma\", \"option_values\": [{\"id\": 1, \"value\": \"8G\"}, {\"id\": 2, \"value\": \"16GB\"}]}, {\"id\": 2, \"name\": \"color\", \"option_values\": [{\"id\": 1, \"value\": \"grey\"}, {\"id\": 2, \"value\": \"pink\"}, {\"id\": 3, \"value\": \"black\"}]}, {\"id\": 3, \"name\": \"brand\", \"option_values\": [{\"id\": 1, \"value\": \"asama\"}, {\"id\": 2, \"value\": \"sony\"}, {\"id\": 3, \"value\": \"jet\"}]}, {\"id\": 4, \"name\": \"battery\", \"option_values\": [{\"id\": 1, \"value\": \"over 5 hour\"}, {\"id\": 2, \"value\": \"over 7 hour\"}]}]}'),(8,'xe-moi-govap-talent-vietduc','{\"attributes\": [{\"id\": 1, \"name\": \"battery\", \"option_values\": [{\"id\": 1, \"value\": \"over 5 hour\"}, {\"id\": 2, \"value\": \"over 7 hour\"}]}, {\"id\": 2, \"name\": \"kilogma\", \"option_values\": [{\"id\": 1, \"value\": \"8G\"}, {\"id\": 2, \"value\": \"16GB\"}]}, {\"id\": 3, \"name\": \"color\", \"option_values\": [{\"id\": 1, \"value\": \"grey\"}, {\"id\": 2, \"value\": \"pink\"}, {\"id\": 3, \"value\": \"black\"}]}, {\"id\": 4, \"name\": \"brand\", \"option_values\": [{\"id\": 1, \"value\": \"asama\"}, {\"id\": 2, \"value\": \"sony\"}, {\"id\": 3, \"value\": \"jet\"}]}]}');
/*!40000 ALTER TABLE `producttypes` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2024-01-08 15:08:37
