CREATE TABLE IF NOT EXISTS `simulation_member` 
(
  `id` VARCHAR(36) CHARACTER SET UTF8MB4 NOT NULL,
  `simulation_id` VARCHAR(36) CHARACTER SET UTF8MB4 NOT NULL, 
  `constructor` ENUM('ALPHA_ROMEO', 'FERRARI', 'HAAS', 'MCLAREN', 'MERCEDES',
        'RACING_POINT', 'RED_BULL_RACING', 'SCUDERIA_TORO_ROSO', 'WILLIAMS') NOT NULL,
  `car_number` INTEGER NOT NULL,
  `force_alarm` BOOLEAN NOT NULL,  
  `no_alarms` BOOLEAN NOT NULL,
  `alarm_occurred` BOOLEAN NULL,
  `alarm_datum_description` ENUM('G_FORCE', 'G_FORCE_DIRECTION', 'FUEL_CONSUMED', 'FUEL_FLOW', 'ENGINE_COOLANT_TEMP',
        'ENGINE_OIL_PRESSURE', 'ENGINE_OIL_TEMP', 'ENGINE_RPM', 'BRAKE_TEMP_FR', 'BRAKE_TEMP_FL',
        'BRAKE_TEMP_RR', 'BRAKE_TEMP_RL', 'ENERGY_STORAGE_LEVEL', 'ENERGY_STORAGE_TEMP', 
        'MGUK_OUTPUT', 'MGUH_OUTPUT', 'SPEED', 'TIRE_PRESSURE_FR', 'TIRE_PRESSURE_FL',
        'TIRE_PRESSURE_RR', 'TIRE_PRESSURE_RL', 'TIRE_TEMP_FR', 'TIRE_TEMP_FL',
        'TIRE_TEMP_RR', 'TIRE_TEMP_RL') NULL,
  `alarm_datum_unit` ENUM('G', 'KG_PER_HOUR', 'DEGREE_CELCIUS', 'MJ', 'JPS',
        'RPM', 'BAR', 'KG', 'KPH', 'METER', 'RADIAN', 'KPA') NULL,
  `alarm_datum_value` FLOAT NULL,
  INDEX par_ind (simulation_id),
  UNIQUE (id, simulation_id),
  CONSTRAINT fk_simulation_id FOREIGN KEY (simulation_id)
  REFERENCES simulation(id)
  ON DELETE CASCADE
  ON UPDATE CASCADE  
) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;