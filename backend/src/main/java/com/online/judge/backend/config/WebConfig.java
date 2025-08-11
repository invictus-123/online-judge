package com.online.judge.backend.config;

import com.online.judge.backend.interceptor.RateLimitInterceptor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;
import org.springframework.lang.NonNull;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class WebConfig implements WebMvcConfigurer {
	private final RateLimitInterceptor rateLimitInterceptor;
	private final String origin;

	public WebConfig(
			RateLimitInterceptor rateLimitInterceptor, 
			@Value("${FRONTEND_HOST_NAME:http://localhost:5173}") String origin) {
		this.rateLimitInterceptor = rateLimitInterceptor;
		this.origin = origin;
	}

	@Override
	public void addInterceptors(@NonNull InterceptorRegistry registry) {
		registry.addInterceptor(rateLimitInterceptor).addPathPatterns("/api/v1/**");
	}

	@Override
	public void addCorsMappings(@NonNull CorsRegistry registry) {
		registry.addMapping("/api/v1/**")
				.allowedOrigins(origin)
				.allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
				.allowedHeaders("*")
				.allowCredentials(true)
				.maxAge(3600);
	}
}
