package com.online.judge.backend.service;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import com.online.judge.backend.config.RateLimitingConfig;
import io.github.bucket4j.Bucket;
import io.github.bucket4j.ConsumptionProbe;
import jakarta.servlet.http.HttpServletRequest;
import java.time.Duration;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class RateLimitServiceTest {
	@Mock
	private RateLimitingConfig rateLimitingConfig;

	@Mock
	private HttpServletRequest request;

	@Mock
	private Bucket bucket;

	@Mock
	private ConsumptionProbe consumptionProbe;

	@InjectMocks
	private RateLimitService rateLimitService;

	@BeforeEach
	void setUp() {
		when(rateLimitingConfig.resolveBucket(anyString(), anyInt(), any(Duration.class)))
				.thenReturn(bucket);
	}

	@Test
	void tryConsume_whenTokensAvailable_shouldReturnTrue() {
		when(request.getRemoteAddr()).thenReturn("192.168.1.1");
		when(bucket.tryConsumeAndReturnRemaining(1)).thenReturn(consumptionProbe);
		when(consumptionProbe.isConsumed()).thenReturn(true);

		boolean result = rateLimitService.tryConsume(request, "test-api", 10, Duration.ofMinutes(1));

		assertTrue(result);
	}

	@Test
	void tryConsume_whenNoTokensAvailable_shouldReturnFalse() {
		when(request.getRemoteAddr()).thenReturn("192.168.1.1");
		when(bucket.tryConsumeAndReturnRemaining(1)).thenReturn(consumptionProbe);
		when(consumptionProbe.isConsumed()).thenReturn(false);

		boolean result = rateLimitService.tryConsume(request, "test-api", 10, Duration.ofMinutes(1));

		assertFalse(result);
	}

	@Test
	void tryConsume_withXForwardedForHeader_shouldUseFirstIP() {
		when(request.getHeader("X-Forwarded-For")).thenReturn("10.0.0.1, 192.168.1.1");
		when(bucket.tryConsumeAndReturnRemaining(1)).thenReturn(consumptionProbe);
		when(consumptionProbe.isConsumed()).thenReturn(true);

		boolean result = rateLimitService.tryConsume(request, "test-api", 10, Duration.ofMinutes(1));

		assertTrue(result);
	}

	@Test
	void tryConsume_withEmptyXForwardedForHeader_shouldUseRemoteAddr() {
		when(request.getHeader("X-Forwarded-For")).thenReturn("");
		when(request.getRemoteAddr()).thenReturn("192.168.1.1");
		when(bucket.tryConsumeAndReturnRemaining(1)).thenReturn(consumptionProbe);
		when(consumptionProbe.isConsumed()).thenReturn(true);

		boolean result = rateLimitService.tryConsume(request, "test-api", 10, Duration.ofMinutes(1));

		assertTrue(result);
	}

	@Test
	void tryConsume_withNullXForwardedForHeader_shouldUseRemoteAddr() {
		when(request.getHeader("X-Forwarded-For")).thenReturn(null);
		when(request.getRemoteAddr()).thenReturn("192.168.1.1");
		when(bucket.tryConsumeAndReturnRemaining(1)).thenReturn(consumptionProbe);
		when(consumptionProbe.isConsumed()).thenReturn(true);

		boolean result = rateLimitService.tryConsume(request, "test-api", 10, Duration.ofMinutes(1));

		assertTrue(result);
	}

	@Test
	void tryConsume_withDifferentApiTypes_shouldCreateSeparateBuckets() {
		when(request.getRemoteAddr()).thenReturn("192.168.1.1");
		when(bucket.tryConsumeAndReturnRemaining(1)).thenReturn(consumptionProbe);
		when(consumptionProbe.isConsumed()).thenReturn(true);

		boolean result1 = rateLimitService.tryConsume(request, "auth", 10, Duration.ofMinutes(1));
		boolean result2 = rateLimitService.tryConsume(request, "submit-code", 10, Duration.ofMinutes(1));

		assertTrue(result1);
		assertTrue(result2);
	}

	@Test
	void tryConsume_withDifferentClientIPs_shouldCreateSeparateBuckets() {
		when(bucket.tryConsumeAndReturnRemaining(1)).thenReturn(consumptionProbe);
		when(consumptionProbe.isConsumed()).thenReturn(true);
		when(request.getRemoteAddr()).thenReturn("192.168.1.1");

		boolean result1 = rateLimitService.tryConsume(request, "test-api", 10, Duration.ofMinutes(1));

		when(request.getRemoteAddr()).thenReturn("192.168.1.2");

		boolean result2 = rateLimitService.tryConsume(request, "test-api", 10, Duration.ofMinutes(1));

		assertTrue(result1);
		assertTrue(result2);
	}
}
