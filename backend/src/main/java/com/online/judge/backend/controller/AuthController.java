package com.online.judge.backend.controller;

import com.online.judge.backend.annotation.RateLimit;
import com.online.judge.backend.dto.request.LoginRequest;
import com.online.judge.backend.dto.request.RegisterRequest;
import com.online.judge.backend.dto.response.AuthResponse;
import com.online.judge.backend.service.UserService;
import com.online.judge.backend.util.JwtUtil;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.validation.Valid;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.web.authentication.logout.SecurityContextLogoutHandler;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/v1/auth")
public class AuthController {
	private static final Logger logger = LoggerFactory.getLogger(AuthController.class);

	private final UserService userService;
	private final AuthenticationManager authenticationManager;
	private final JwtUtil jwtUtil;

	public AuthController(UserService userService, AuthenticationManager authenticationManager, JwtUtil jwtUtil) {
		this.userService = userService;
		this.authenticationManager = authenticationManager;
		this.jwtUtil = jwtUtil;
	}

	@PostMapping("/register")
	@RateLimit(apiType = "auth", capacity = 10, refillPeriodMinutes = 1)
	public ResponseEntity<AuthResponse> registerUser(@Valid @RequestBody RegisterRequest registerRequest) {
		logger.info(
				"Received registration request for user: {} with role: {}",
				registerRequest.handle(),
				registerRequest.userRole());

		userService.registerUser(registerRequest);

		logger.info(
				"User registered successfully with handle: {}. Authenticating the user...", registerRequest.handle());

		Authentication authentication = authenticationManager.authenticate(
				new UsernamePasswordAuthenticationToken(registerRequest.handle(), registerRequest.password()));

		final UserDetails userDetails = (UserDetails) authentication.getPrincipal();
		final String token = jwtUtil.generateToken(userDetails);

		logger.info("User authenticated successfully with handle: {}", registerRequest.handle());
		return ResponseEntity.ok(new AuthResponse(token));
	}

	@PostMapping("/login")
	@RateLimit(apiType = "auth", capacity = 10, refillPeriodMinutes = 1)
	public ResponseEntity<AuthResponse> loginUser(@Valid @RequestBody LoginRequest loginRequest) {
		logger.info("Attempting to authenticate user with handle: {}", loginRequest.handle());

		Authentication authentication = authenticationManager.authenticate(
				new UsernamePasswordAuthenticationToken(loginRequest.handle(), loginRequest.password()));

		final UserDetails userDetails = (UserDetails) authentication.getPrincipal();
		final String token = jwtUtil.generateToken(userDetails);

		logger.info("User authenticated successfully with handle: {}", loginRequest.handle());
		return ResponseEntity.ok(new AuthResponse(token));
	}

	@PostMapping("/logout")
	public ResponseEntity<String> logoutUser(HttpServletRequest request, HttpServletResponse response) {
		logger.info("Received request to log out user");

		Authentication auth = SecurityContextHolder.getContext().getAuthentication();
		if (auth != null) {
			logger.info("Logging out user: {}", auth.getName());
			new SecurityContextLogoutHandler().logout(request, response, auth);
		}
		return ResponseEntity.ok("User logged out successfully");
	}
}
