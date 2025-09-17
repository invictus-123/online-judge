package com.online.judge.executorv2.controller;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.time.LocalDateTime;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/executor")
public class HealthController {

	@GetMapping("/status")
	public Map<String, Object> getStatus() {
		return Map.of(
			"service", "executorv2",
			"status", "running",
			"timestamp", LocalDateTime.now(),
			"version", "0.0.1-SNAPSHOT"
		);
	}
}