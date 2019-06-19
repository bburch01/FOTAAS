CREATE TABLE IF NOT EXISTS `telemetry_datum` 
(
  `id` VARCHAR(36) CHARACTER SET UTF8MB4 NOT NULL,
  `simulated` BOOLEAN NOT NULL,
  `simulation_id` VARCHAR(36) CHARACTER SET UTF8MB4,
  `simulation_transmit_sequence_number` INTEGER NOT NULL,
  `grand_prix` ENUM('UNITED_STATES', 'AZERBAIJAN', 'SPANISH', 'GERMAN', 'HUNGARIAN',
        'BRAZILIAN', 'SINGAPORE', 'AUSTRALIAN', 'MEXICAN', 'MONACO',
        'CANADIAN', 'ITALIAN', 'FRENCH', 'BAHRAIN', 'CHINESE',
        'BRITISH', 'RUSSIAN', 'BELGIAN', 'AUSTRIAN', 'JAPANESE', 'ABU_DHABI') NOT NULL,
  `track` ENUM('AUSTIN', 'BAKU', 'CATALUNYA_BARCELONA', 'HOCKENHEIM', 'HUNGARORING',
        'INTERLAGOS_SAU_PAULO', 'MARINA_BAY', 'MELBOURNE', 'MEXICO_CITY', 'MONTE_CARLO',
        'MONTREAL', 'MONZA', 'PAUL_RICARD_LE_CASTELLET', 'SAKHIR', 'SHANGHAI',
        'SILVERSTONE', 'SOCHI', 'SPA_FRANCORCHAMPS', 'SPIELBERG_RED_BULL_RING', 'SUZUKA', 'YAS_MARINA') NOT NULL,
  `constructor` ENUM('ALPHA_ROMEO', 'FERRARI', 'HAAS', 'MCLAREN', 'MERCEDES',
        'RACING_POINT', 'RED_BULL_RACING', 'SCUDERIA_TORO_ROSO', 'WILLIAMS') NOT NULL,
  `car_number` INTEGER NOT NULL,
  `timestamp` TIMESTAMP NOT NULL,
  `latitude` INTEGER NOT NULL,
  `longitude` INTEGER NOT NULL,
  `elevation` INTEGER NOT NULL,
  `description` ENUM('G_FORCE', 'G_FORCE_DIRECTION', 'FUEL_CONSUMED', 'FUEL_FLOW', 'ENGINE_COOLANT_TEMP',
        'ENGINE_OIL_PRESSURE', 'ENGINE_OIL_TEMP', 'ENGINE_RPM', 'BRAKE_TEMP_FR', 'BRAKE_TEMP_FL',
        'BRAKE_TEMP_RR', 'BRAKE_TEMP_RL', 'ENERGY_STORAGE_LEVEL', 'ENERGY_STORAGE_TEMP', 
        'MGUK_OUTPUT', 'MGUH_OUTPUT', 'SPEED', 'TIRE_PRESSURE_FR', 'TIRE_PRESSURE_FL',
        'TIRE_PRESSURE_RR', 'TIRE_PRESSURE_RL', 'TIRE_TEMP_FR', 'TIRE_TEMP_FL',
        'TIRE_TEMP_RR', 'TIRE_TEMP_RL') NOT NULL,
  `unit` ENUM('G', 'KG_PER_HOUR', 'DEGREE_CELCIUS', 'MJ', 'JPS',
        'RPM', 'BAR', 'KG', 'KPH', 'METER', 'RADIAN', 'KPA') NOT NULL,
  `value` FLOAT NOT NULL,
  `hi_alarm` BOOLEAN NOT NULL,  
  `lo_alarm` BOOLEAN NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;