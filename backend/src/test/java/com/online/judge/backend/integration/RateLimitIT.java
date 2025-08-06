package com.online.judge.backend.integration;

import static org.assertj.core.api.Assertions.assertThat;

import com.online.judge.backend.dto.request.RegisterRequest;
import com.online.judge.backend.dto.response.AuthResponse;
import com.online.judge.backend.model.shared.UserRole;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

/**
 * Integration test for rate limiting functionality.
 * Uses the 'integration' profile which has increased rate limits (100 requests/minute).
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@Testcontainers
@Tag("integration")
@ActiveProfiles("integration")
class RateLimitIT {
	@Container
	private static final PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:15-alpine");

	@DynamicPropertySource
	private static void configureProperties(DynamicPropertyRegistry registry) {
		registry.add("spring.datasource.url", postgres::getJdbcUrl);
		registry.add("spring.datasource.username", postgres::getUsername);
		registry.add("spring.datasource.password", postgres::getPassword);
		registry.add("spring.datasource.driver-class-name", () -> "org.postgresql.Driver");
		registry.add("spring.jpa.hibernate.ddl-auto", () -> "create");
	}

	@Autowired
	private TestRestTemplate restTemplate;

	@Test
	void authEndpoint_shouldAllow100RequestsPerMinute() {
		String baseUrl = "/api/v1/auth/register";
		AtomicInteger successCount = new AtomicInteger(0);
		AtomicInteger rateLimitedCount = new AtomicInteger(0);

		ExecutorService executor = Executors.newFixedThreadPool(20);
		int numRequests = 50;
		CompletableFuture<Void>[] futures = new CompletableFuture[numRequests];

		for (int i = 0; i < numRequests; i++) {
			final int requestIndex = i;
			futures[i] = CompletableFuture.runAsync(
					() -> {
						String userHandle = "testuser" + requestIndex + "-"
								+ UUID.randomUUID().toString().substring(0, 8);
						String userEmail = "testuser" + requestIndex + "-" + UUID.randomUUID() + "@example.com";
						RegisterRequest registerRequest = new RegisterRequest(
								userHandle, userEmail, "password123", "Test", "User", UserRole.USER);

						ResponseEntity<AuthResponse> response =
								restTemplate.postForEntity(baseUrl, registerRequest, AuthResponse.class);

						if (response.getStatusCode() == HttpStatus.OK) {
							successCount.incrementAndGet();
						} else if (response.getStatusCode() == HttpStatus.TOO_MANY_REQUESTS) {
							rateLimitedCount.incrementAndGet();
						}
					},
					executor);
		}

		CompletableFuture.allOf(futures).join();
		executor.shutdown();

		assertThat(successCount.get()).isEqualTo(numRequests);
		assertThat(rateLimitedCount.get()).isZero();

		System.out.printf(
				"Completed authEndpoint_shouldAllow100RequestsPerMinute test with %d success and %d rate limited requests.",
				successCount.get(), rateLimitedCount.get());
	}

	@Test
	void authEndpoint_shouldAllow100RequestsPerMinute_moreThan100Requests() {
		String baseUrl = "/api/v1/auth/register";
		AtomicInteger successCount = new AtomicInteger(0);
		AtomicInteger rateLimitedCount = new AtomicInteger(0);

		ExecutorService executor = Executors.newFixedThreadPool(20);
		int numRequests = 120;
		CompletableFuture<Void>[] futures = new CompletableFuture[numRequests];

		for (int i = 0; i < numRequests; i++) {
			final int requestIndex = i;
			futures[i] = CompletableFuture.runAsync(
					() -> {
						String userHandle = "testuser" + requestIndex + "-"
								+ UUID.randomUUID().toString().substring(0, 8);
						String userEmail = "testuser" + requestIndex + "-" + UUID.randomUUID() + "@example.com";
						RegisterRequest registerRequest = new RegisterRequest(
								userHandle, userEmail, "password123", "Test", "User", UserRole.USER);

						ResponseEntity<AuthResponse> response =
								restTemplate.postForEntity(baseUrl, registerRequest, AuthResponse.class);

						if (response.getStatusCode() == HttpStatus.OK) {
							successCount.incrementAndGet();
						} else if (response.getStatusCode() == HttpStatus.TOO_MANY_REQUESTS) {
							rateLimitedCount.incrementAndGet();
						}
					},
					executor);
		}

		CompletableFuture.allOf(futures).join();
		executor.shutdown();

		assertThat(successCount.get()).isLessThanOrEqualTo(100);
		assertThat(rateLimitedCount.get()).isGreaterThan(0);

		System.out.printf(
				"Completed authEndpoint_shouldAllow100RequestsPerMinute_moreThan100Requests test with %d success and %d rate limited requests.",
				successCount.get(), rateLimitedCount.get());
	}
}
