package com.online.judge.backend.service;

import com.online.judge.backend.config.RateLimitingConfig;
import io.github.bucket4j.Bucket;
import io.github.bucket4j.ConsumptionProbe;
import jakarta.servlet.http.HttpServletRequest;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.time.Duration;
import org.springframework.stereotype.Service;

@Service
public class RateLimitService {
	private final RateLimitingConfig rateLimitingConfig;

	public RateLimitService(RateLimitingConfig rateLimitingConfig) {
		this.rateLimitingConfig = rateLimitingConfig;
	}

	public boolean tryConsume(HttpServletRequest request, String apiType, int capacity, Duration refillPeriod) {
		String clientIdHash = getClientIdHash(request);
		String key = apiType + ":" + clientIdHash;

		Bucket bucket = rateLimitingConfig.resolveBucket(key, capacity, refillPeriod);
		ConsumptionProbe probe = bucket.tryConsumeAndReturnRemaining(1);

		return probe.isConsumed();
	}

	private String getClientIdHash(HttpServletRequest request) {
		String clientId = getClientId(request);
		return hashClientId(clientId);
	}

	private String getClientId(HttpServletRequest request) {
		String xForwardedFor = request.getHeader("X-Forwarded-For");
		if (xForwardedFor != null && !xForwardedFor.isEmpty()) {
			return xForwardedFor.split(",")[0].trim();
		}
		return request.getRemoteAddr();
	}

	private String hashClientId(String clientId) {
		try {
			MessageDigest digest = MessageDigest.getInstance("SHA-256");
			byte[] hash = digest.digest(clientId.getBytes(StandardCharsets.UTF_8));
			StringBuilder hexString = new StringBuilder();
			for (byte b : hash) {
				String hex = Integer.toHexString(0xff & b);
				if (hex.length() == 1) {
					hexString.append('0');
				}
				hexString.append(hex);
			}
			// Use only first 16 characters to reduce memory footprint
			return hexString.toString().substring(0, 16);
		} catch (NoSuchAlgorithmException e) {
			// Fallback to simple hash if SHA-256 is not available (unlikely)
			return String.valueOf(clientId.hashCode());
		}
	}
}
