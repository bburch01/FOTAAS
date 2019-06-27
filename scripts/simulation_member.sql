CREATE TABLE IF NOT EXISTS `simulation_member` 
(
  `id` VARCHAR(36) CHARACTER SET UTF8MB4 NOT NULL,
  `simulation_id` VARCHAR(36) CHARACTER SET UTF8MB4 NOT NULL, 
  `constructor` ENUM('ALPHA_ROMEO', 'FERRARI', 'HAAS', 'MCLAREN', 'MERCEDES',
        'RACING_POINT', 'RED_BULL_RACING', 'SCUDERIA_TORO_ROSO', 'WILLIAMS') NOT NULL,
  `car_number` INTEGER NOT NULL,
  `force_alarm` BOOLEAN NOT NULL,  
  `no_alarms` BOOLEAN NOT NULL,
  INDEX par_ind (simulation_id),
  UNIQUE (id, simulation_id),
  CONSTRAINT fk_simulation_id FOREIGN KEY (simulation_id)
  REFERENCES simulation(id)
  ON DELETE CASCADE
  ON UPDATE CASCADE  
) ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4;