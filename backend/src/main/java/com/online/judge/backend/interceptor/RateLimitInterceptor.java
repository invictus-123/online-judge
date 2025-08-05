package com.online.judge.backend.interceptor;

import com.online.judge.backend.annotation.RateLimit;
import com.online.judge.backend.service.RateLimitConfigService;
import com.online.judge.backend.service.RateLimitService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.time.Duration;
import org.springframework.http.HttpStatus;
import org.springframework.lang.NonNull;
import org.springframework.stereotype.Component;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.HandlerInterceptor;

@Component
public class RateLimitInterceptor implements HandlerInterceptor {
	private final RateLimitConfigService rateLimitConfigService;
	private final RateLimitService rateLimitService;

	public RateLimitInterceptor(RateLimitConfigService rateLimitConfigService, RateLimitService rateLimitService) {
		this.rateLimitConfigService = rateLimitConfigService;
		this.rateLimitService = rateLimitService;
	}

	@Override
	public boolean preHandle(
			@NonNull HttpServletRequest request, @NonNull HttpServletResponse response, @NonNull Object handler)
			throws Exception {
		if (!(handler instanceof HandlerMethod)) {
			return true;
		}

		HandlerMethod handlerMethod = (HandlerMethod) handler;
		RateLimit rateLimit = handlerMethod.getMethodAnnotation(RateLimit.class);
		if (rateLimit == null) {
			return true;
		}

		Duration refillPeriod = Duration.ofMinutes(rateLimit.refillPeriodMinutes());
		int effectiveCapacity = rateLimitConfigService.getEffectiveCapacity(rateLimit.capacity());
		boolean allowed = rateLimitService.tryConsume(request, rateLimit.apiType(), effectiveCapacity, refillPeriod);
		if (!allowed) {
			response.setStatus(HttpStatus.TOO_MANY_REQUESTS.value());
			response.getWriter().write("{\"error\":\"Too many requests. Please try again later.\"}");
			response.setContentType("application/json");
			return false;
		}

		return true;
	}
}
