-- Minimal schema for bobby tests
-- Only includes tables needed for testing
-- Foreign keys are one-way only (from organization_plans to other tables) to avoid circular dependencies

CREATE TABLE `organizations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `identifier` varchar(255) NOT NULL,
  `role` enum('RL_1','RL_2') NOT NULL DEFAULT 'RL_1',
  `is_authed` tinyint(1) NOT NULL DEFAULT 0,
  `is_deleted` tinyint(1) NOT NULL DEFAULT 0,
  `deleted_at` datetime(6) DEFAULT NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_organizations_identifier` (`identifier`),
  KEY `id_organizations_role` (`role`),
  KEY `id_organizations_is_authed` (`is_authed`),
  KEY `id_organizations_is_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `banners` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '',
  `state` enum(
    "BS_1",
    "BS_2",
    "BS_3",
    "BS_4",
    "BS_5",
    "BS_6",
    "BS_7",
    "BS_8",
    "BS_9"
  ) NOT NULL DEFAULT 'BS_1',
  `sort` int(11) NOT NULL DEFAULT 0,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  KEY `id_banners_state` (`state`),
  KEY `id_banners_sort` (`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `organization_plan_buyers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `company_name` varchar(255) NOT NULL,
  `phone_number` varchar(128) NOT NULL,
  `email` varchar(255) NOT NULL,
  `representative` varchar(255) NOT NULL,
  `user_name` varchar(255) NOT NULL,
  `postcode` varchar(255) NOT NULL,
  `address1` varchar(255) NOT NULL,
  `address2` varchar(255) NOT NULL,
  `invoice_postcode` varchar(255) NOT NULL DEFAULT '',
  `invoice_address1` varchar(255) NOT NULL DEFAULT '',
  `invoice_address2` varchar(255) NOT NULL DEFAULT '',
  `invoice_user_name` varchar(255) NOT NULL DEFAULT '',
  `invoice_email` varchar(255) NOT NULL DEFAULT '',
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `organization_plans` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `organization_id` int(11) NOT NULL,
  `organization_plan_buyer_id` int(11) NOT NULL,
  `plan` enum(
    'PL_1',
    'PL_2',
    'PL_3',
    'PL_4',
    'PL_5',
    'PL_6',
    'PL_7',
    'PL_8',
    'PL_9',
    'PL_10',
    'PL_11',
    'PL_12',
    'PL_13',
    'PL_14',
    'PL_15',
    'PL_16',
    'PL_17',
    'PL_18',
    'PL_19',
    'PL_20',
    'PL_21'
  ) NOT NULL DEFAULT 'PL_1',
  `payment_method` enum('PM_1', 'PM_2') NOT NULL DEFAULT 'PM_2',
  `stripe_customer` varchar(255) NOT NULL DEFAULT '',
  `stripe_subscription` varchar(255) NOT NULL DEFAULT '',
  `stripe_charge` varchar(255) NOT NULL DEFAULT '',
  `unsubscribed_reason` varchar(2048) NOT NULL DEFAULT '',
  `unsubscribed_requested_at` datetime(6) NULL DEFAULT NULL,
  `unsubscribed_actually_at` datetime(6) NULL DEFAULT NULL,
  `is_pending` tinyint(1) NOT NULL DEFAULT 0,
  `is_enabled` tinyint(1) DEFAULT NULL,
  `contract_price` int(11) NOT NULL DEFAULT 0,
  `contract_version` int(11) NOT NULL,
  `contract_period` int(11) NOT NULL,
  `contract_updated_at` datetime(6) NULL DEFAULT NULL,
  `contract_started_at` datetime(6) NULL DEFAULT NULL,
  `contract_expired_at` datetime(6) NOT NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  KEY `id_organization_plans_is_pending` (`is_pending`),
  KEY `id_organization_plans_is_enabled` (`is_enabled`),
  KEY `id_organization_plans_organization_id` (`organization_id`),
  KEY `id_organization_plans_organization_plan_buyer_id` (`organization_plan_buyer_id`),
  UNIQUE KEY `uk_organization_plan_organization_id_is_enabled` (`organization_id`,`is_enabled`),
  CONSTRAINT `fk_organization_plans_organization_id` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`),
  CONSTRAINT `fk_organization_plans_organization_plan_buyer_id` FOREIGN KEY (`organization_plan_buyer_id`) REFERENCES `organization_plan_buyers` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Insert test organizations for testing
INSERT INTO `organizations` (`id`, `identifier`, `role`, `is_authed`) VALUES
(500, 'test-org-500', 'RL_1', 1),
(501, 'test-org-501', 'RL_1', 1),
(502, 'test-org-502', 'RL_1', 1),
(503, 'test-org-503', 'RL_1', 1),
(504, 'test-org-504', 'RL_1', 1),
(505, 'test-org-505', 'RL_1', 1),
(506, 'test-org-506', 'RL_1', 1),
(507, 'test-org-507', 'RL_1', 1),
(508, 'test-org-508', 'RL_1', 1),
(509, 'test-org-509', 'RL_1', 1),
(510, 'test-org-510', 'RL_1', 1),
(511, 'test-org-511', 'RL_1', 1),
(512, 'test-org-512', 'RL_1', 1),
(513, 'test-org-513', 'RL_1', 1),
(514, 'test-org-514', 'RL_1', 1),
(515, 'test-org-515', 'RL_1', 1),
(516, 'test-org-516', 'RL_1', 1),
(517, 'test-org-517', 'RL_1', 1),
(518, 'test-org-518', 'RL_1', 1),
(519, 'test-org-519', 'RL_1', 1),
(520, 'test-org-520', 'RL_1', 1);
