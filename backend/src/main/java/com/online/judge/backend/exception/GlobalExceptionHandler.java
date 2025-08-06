package com.online.judge.backend.exception;

import com.online.judge.backend.dto.response.ErrorResponse;
import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.AccessDeniedException;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.authentication.DisabledException;
import org.springframework.security.authentication.LockedException;
import org.springframework.validation.BindException;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.method.annotation.MethodArgumentTypeMismatchException;

@RestControllerAdvice
public class GlobalExceptionHandler {
	private static final Logger logger = LoggerFactory.getLogger(GlobalExceptionHandler.class);

	@ExceptionHandler(ProblemNotFoundException.class)
	public ResponseEntity<ErrorResponse> handleProblemNotFound(
			ProblemNotFoundException ex, HttpServletRequest request) {
		logger.warn("Problem not found: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				ex.getMessage(), "PROBLEM_NOT_FOUND", HttpStatus.NOT_FOUND.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.NOT_FOUND).body(errorResponse);
	}

	@ExceptionHandler(SubmissionNotFoundException.class)
	public ResponseEntity<ErrorResponse> handleSubmissionNotFound(
			SubmissionNotFoundException ex, HttpServletRequest request) {
		logger.warn("Submission not found: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				ex.getMessage(), "SUBMISSION_NOT_FOUND", HttpStatus.NOT_FOUND.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.NOT_FOUND).body(errorResponse);
	}

	@ExceptionHandler(UserAlreadyExistsException.class)
	public ResponseEntity<ErrorResponse> handleUserAlreadyExists(
			UserAlreadyExistsException ex, HttpServletRequest request) {
		logger.warn("User already exists: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				ex.getMessage(), "USER_ALREADY_EXISTS", HttpStatus.CONFLICT.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.CONFLICT).body(errorResponse);
	}

	@ExceptionHandler(UserNotAuthorizedException.class)
	public ResponseEntity<ErrorResponse> handleUserNotAuthorized(
			UserNotAuthorizedException ex, HttpServletRequest request) {
		logger.warn("User not authorized: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				ex.getMessage(), "UNAUTHORIZED", HttpStatus.FORBIDDEN.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.FORBIDDEN).body(errorResponse);
	}

	@ExceptionHandler(BadCredentialsException.class)
	public ResponseEntity<ErrorResponse> handleBadCredentials(BadCredentialsException ex, HttpServletRequest request) {
		logger.warn("Bad credentials: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				"Invalid username or password",
				"BAD_CREDENTIALS",
				HttpStatus.UNAUTHORIZED.value(),
				request.getRequestURI());
		return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(errorResponse);
	}

	@ExceptionHandler(DisabledException.class)
	public ResponseEntity<ErrorResponse> handleDisabled(DisabledException ex, HttpServletRequest request) {
		logger.warn("Account disabled: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				"Account is disabled", "ACCOUNT_DISABLED", HttpStatus.FORBIDDEN.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.FORBIDDEN).body(errorResponse);
	}

	@ExceptionHandler(LockedException.class)
	public ResponseEntity<ErrorResponse> handleLocked(LockedException ex, HttpServletRequest request) {
		logger.warn("Account locked: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				"Account is locked", "ACCOUNT_LOCKED", HttpStatus.FORBIDDEN.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.FORBIDDEN).body(errorResponse);
	}

	@ExceptionHandler(AccessDeniedException.class)
	public ResponseEntity<ErrorResponse> handleAccessDenied(AccessDeniedException ex, HttpServletRequest request) {
		logger.warn("Access denied: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				"Access denied", "ACCESS_DENIED", HttpStatus.FORBIDDEN.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.FORBIDDEN).body(errorResponse);
	}

	@ExceptionHandler({MethodArgumentNotValidException.class})
	public ResponseEntity<ErrorResponse> handleValidationExceptions(
			MethodArgumentNotValidException ex, HttpServletRequest request) {
		logger.warn("Validation failed: {}", ex.getMessage());

		List<ErrorResponse.FieldError> fieldErrors = ex.getBindingResult().getFieldErrors().stream()
				.map(fieldError -> new ErrorResponse.FieldError(
						fieldError.getField(), fieldError.getRejectedValue(), fieldError.getDefaultMessage()))
				.toList();

		ErrorResponse errorResponse = new ErrorResponse(
				"Validation failed",
				"VALIDATION_ERROR",
				HttpStatus.BAD_REQUEST.value(),
				request.getRequestURI(),
				fieldErrors);
		return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(errorResponse);
	}

	@ExceptionHandler(BindException.class)
	public ResponseEntity<ErrorResponse> handleBindException(BindException ex, HttpServletRequest request) {
		logger.warn("Binding failed: {}", ex.getMessage());

		List<ErrorResponse.FieldError> fieldErrors = ex.getBindingResult().getFieldErrors().stream()
				.map(fieldError -> new ErrorResponse.FieldError(
						fieldError.getField(), fieldError.getRejectedValue(), fieldError.getDefaultMessage()))
				.toList();

		ErrorResponse errorResponse = new ErrorResponse(
				"Request binding failed",
				"BINDING_ERROR",
				HttpStatus.BAD_REQUEST.value(),
				request.getRequestURI(),
				fieldErrors);
		return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(errorResponse);
	}

	@ExceptionHandler(MethodArgumentTypeMismatchException.class)
	public ResponseEntity<ErrorResponse> handleTypeMismatch(
			MethodArgumentTypeMismatchException ex, HttpServletRequest request) {
		logger.warn("Type mismatch: {}", ex.getMessage());
		String message = String.format("Invalid value '%s' for parameter '%s'", ex.getValue(), ex.getName());
		ErrorResponse errorResponse =
				new ErrorResponse(message, "TYPE_MISMATCH", HttpStatus.BAD_REQUEST.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(errorResponse);
	}

	@ExceptionHandler(DataIntegrityViolationException.class)
	public ResponseEntity<ErrorResponse> handleDataIntegrityViolation(
			DataIntegrityViolationException ex, HttpServletRequest request) {
		logger.error("Data integrity violation: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				"Data integrity constraint violation",
				"DATA_INTEGRITY_ERROR",
				HttpStatus.CONFLICT.value(),
				request.getRequestURI());
		return ResponseEntity.status(HttpStatus.CONFLICT).body(errorResponse);
	}

	@ExceptionHandler(IllegalArgumentException.class)
	public ResponseEntity<ErrorResponse> handleIllegalArgument(
			IllegalArgumentException ex, HttpServletRequest request) {
		logger.warn("Illegal argument: {}", ex.getMessage());
		ErrorResponse errorResponse = new ErrorResponse(
				ex.getMessage(), "ILLEGAL_ARGUMENT", HttpStatus.BAD_REQUEST.value(), request.getRequestURI());
		return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(errorResponse);
	}

	@ExceptionHandler(Exception.class)
	public ResponseEntity<ErrorResponse> handleGenericException(Exception ex, HttpServletRequest request) {
		logger.error("Unexpected error occurred", ex);
		ErrorResponse errorResponse = new ErrorResponse(
				"An unexpected error occurred",
				"INTERNAL_SERVER_ERROR",
				HttpStatus.INTERNAL_SERVER_ERROR.value(),
				request.getRequestURI());
		return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(errorResponse);
	}
}
