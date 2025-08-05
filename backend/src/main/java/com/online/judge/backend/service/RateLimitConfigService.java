package com.online.judge.backend.service;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

@Service
public class RateLimitConfigService {
	@Value("${rate-limit.multiplier:1}")
	private int multiplier;

	public int getEffectiveCapacity(int baseCapacity) {
		return baseCapacity * multiplier;
	}
}
