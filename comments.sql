CREATE TABLE comments (
    `comment_id` int NOT NULL AUTO_INCREMENT,
    `comment_on` varchar(64) NOT NULL,
    `comment` varchar(1028) NOT NULL,
    `email` varchar(128),
    `at` datetime NOT NULL,
    `addr` varchar(64) DEFAULT NULL,
    `client_ip` varchar(64) DEFAULT NULL,
    `x_forwarded_for` varchar(64) DEFAULT NULL,
    `port` varchar(8) DEFAULT NULL,
    `ua` varchar(256) DEFAULT NULL,
    `referer` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`comment_id`)
)

