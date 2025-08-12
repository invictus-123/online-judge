package com.online.judge.backend.config;

import com.online.judge.backend.service.CustomUserDetailsService;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.AuthenticationProvider;
import org.springframework.security.authentication.dao.DaoAuthenticationProvider;
import org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.HeadersConfigurer.FrameOptionsConfig;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.cors.CorsConfigurationSource;

/** Security configuration class. */
@Configuration
@EnableWebSecurity
public class SecurityConfig {
	private static final String[] PUBLIC_ENDPOINTS = new String[] {
		"/h2-console/**",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/problems/{id}",
		"/api/v1/problems/list",
		"/api/v1/submissions/list",
		"/api/v1/submissions/{id}",
		"/swagger-ui.html",
		"/swagger-ui/**",
		"/v3/api-docs/**"
	};

	private final JwtAuthFilter jwtAuthFilter;
	private final CorsConfigurationSource corsConfigurationSource;
	private final CustomUserDetailsService userDetailsService;

	public SecurityConfig(
			JwtAuthFilter jwtAuthFilter,
			CorsConfigurationSource corsConfigurationSource,
			CustomUserDetailsService userDetailsService) {
		this.jwtAuthFilter = jwtAuthFilter;
		this.corsConfigurationSource = corsConfigurationSource;
		this.userDetailsService = userDetailsService;
	}

	/**
	 * Creates a PasswordEncoder bean that can be injected into other components.
	 *
	 * @return An instance of BCryptPasswordEncoder.
	 */
	@Bean
	public PasswordEncoder passwordEncoder() {
		return new BCryptPasswordEncoder();
	}

	/**
	 * Configures the security filter chain for the application.
	 *
	 * @param http
	 *            The HttpSecurity object to configure.
	 * @return The configured SecurityFilterChain.
	 * @throws Exception
	 *             If an error occurs during configuration.
	 */
	@Bean
	public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
		return http.csrf(csrf -> csrf.disable())
				.cors(cors -> cors.configurationSource(corsConfigurationSource))
				.authorizeHttpRequests(auth -> auth.requestMatchers(PUBLIC_ENDPOINTS)
						.permitAll()
						.anyRequest()
						.authenticated())
				.headers(headers -> headers.frameOptions(FrameOptionsConfig::sameOrigin))
				.sessionManagement(session -> session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
				.authenticationProvider(authenticationProvider())
				.addFilterBefore(jwtAuthFilter, UsernamePasswordAuthenticationFilter.class)
				.build();
	}

	/**
	 * Creates an AuthenticationProvider bean that uses the UserDetailsService and
	 * PasswordEncoder.
	 *
	 * @return An instance of DaoAuthenticationProvider.
	 */
	@Bean
	public AuthenticationProvider authenticationProvider() {
		DaoAuthenticationProvider authProvider = new DaoAuthenticationProvider();
		authProvider.setUserDetailsService(userDetailsService);
		authProvider.setPasswordEncoder(passwordEncoder());
		return authProvider;
	}

	/**
	 * Creates an AuthenticationManager bean that can be used for authentication.
	 *
	 * @param config
	 *            The AuthenticationConfiguration object.
	 * @return An instance of AuthenticationManager.
	 * @throws Exception
	 *             If an error occurs during creation.
	 */
	@Bean
	public AuthenticationManager authenticationManager(AuthenticationConfiguration config) throws Exception {
		return config.getAuthenticationManager();
	}
}
