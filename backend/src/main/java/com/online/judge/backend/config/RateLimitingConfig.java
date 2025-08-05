package com.online.judge.backend.config;

import io.github.bucket4j.Bandwidth;
import io.github.bucket4j.Bucket;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import java.time.Duration;
import java.time.Instant;
import java.util.Map.Entry;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RateLimitingConfig {
	private static final int MAX_CACHE_SIZE = 10000;
	private static final int CLEANUP_INTERVAL_MINUTES = 10;

	private final ConcurrentHashMap<String, BucketWrapper> cache = new ConcurrentHashMap<>();
	private ScheduledExecutorService cleanupExecutor;

	@PostConstruct
	public void init() {
		cleanupExecutor = Executors.newSingleThreadScheduledExecutor();
		cleanupExecutor.scheduleWithFixedDelay(
				this::cleanupExpiredBuckets, CLEANUP_INTERVAL_MINUTES, CLEANUP_INTERVAL_MINUTES, TimeUnit.MINUTES);
	}

	@PreDestroy
	public void destroy() {
		if (cleanupExecutor != null) {
			cleanupExecutor.shutdown();
		}
	}

	@Bean
	public ConcurrentMap<String, BucketWrapper> rateLimitCache() {
		return cache;
	}

	public Bucket createNewBucket(String key, int capacity, Duration refillPeriod) {
		if (cache.size() >= MAX_CACHE_SIZE) {
			removeOldestEntries();
		}

		Bandwidth limit = Bandwidth.builder()
				.capacity(capacity)
				.refillIntervally(capacity, refillPeriod)
				.build();
		Bucket bucket = Bucket.builder().addLimit(limit).build();
		BucketWrapper wrapper = new BucketWrapper(bucket, Instant.now());
		cache.put(key, wrapper);
		return bucket;
	}

	public Bucket resolveBucket(String key, int capacity, Duration refillPeriod) {
		BucketWrapper wrapper = cache.computeIfAbsent(key, k -> {
			Bandwidth limit = Bandwidth.builder()
					.capacity(capacity)
					.refillIntervally(capacity, refillPeriod)
					.build();
			Bucket bucket = Bucket.builder().addLimit(limit).build();
			return new BucketWrapper(bucket, Instant.now());
		});

		wrapper.updateLastAccess();
		return wrapper.getBucket();
	}

	private void cleanupExpiredBuckets() {
		Instant cutoff = Instant.now().minus(Duration.ofHours(1));
		cache.entrySet().removeIf(entry -> entry.getValue().getLastAccess().isBefore(cutoff));
	}

	private void removeOldestEntries() {
		int toRemove = MAX_CACHE_SIZE / 10;
		cache.entrySet().stream()
				.sorted((e1, e2) ->
						e1.getValue().getLastAccess().compareTo(e2.getValue().getLastAccess()))
				.limit(toRemove)
				.map(Entry::getKey)
				.forEach(cache::remove);
	}

	private static class BucketWrapper {
		private final Bucket bucket;
		private volatile Instant lastAccess;

		public BucketWrapper(Bucket bucket, Instant lastAccess) {
			this.bucket = bucket;
			this.lastAccess = lastAccess;
		}

		public Bucket getBucket() {
			return bucket;
		}

		public Instant getLastAccess() {
			return lastAccess;
		}

		public void updateLastAccess() {
			this.lastAccess = Instant.now();
		}
	}
}
