package com.online.judge.backend.exception;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import com.online.judge.backend.dto.response.ErrorResponse;
import jakarta.servlet.http.HttpServletRequest;
import java.util.Collections;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.AccessDeniedException;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.validation.BindingResult;
import org.springframework.validation.FieldError;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.method.annotation.MethodArgumentTypeMismatchException;

class GlobalExceptionHandlerTest {
	private GlobalExceptionHandler exceptionHandler;

	@Mock
	private HttpServletRequest request;

	@Mock
	private MethodArgumentNotValidException validationException;

	@Mock
	private BindingResult bindingResult;

	@BeforeEach
	void setUp() {
		MockitoAnnotations.openMocks(this);
		exceptionHandler = new GlobalExceptionHandler();
		when(request.getRequestURI()).thenReturn("/api/v1/test");
	}

	@Test
	void testHandleProblemNotFoundException() {
		ProblemNotFoundException ex = new ProblemNotFoundException("Problem not found");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleProblemNotFound(ex, request);

		assertEquals(HttpStatus.NOT_FOUND, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("Problem not found", response.getBody().message());
		assertEquals("PROBLEM_NOT_FOUND", response.getBody().error());
		assertEquals(404, response.getBody().status());
		assertEquals("/api/v1/test", response.getBody().path());
	}

	@Test
	void testHandleSubmissionNotFoundException() {
		SubmissionNotFoundException ex = new SubmissionNotFoundException("Submission not found");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleSubmissionNotFound(ex, request);

		assertEquals(HttpStatus.NOT_FOUND, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("Submission not found", response.getBody().message());
		assertEquals("SUBMISSION_NOT_FOUND", response.getBody().error());
	}

	@Test
	void testHandleUserAlreadyExistsException() {
		UserAlreadyExistsException ex = new UserAlreadyExistsException("User already exists");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleUserAlreadyExists(ex, request);

		assertEquals(HttpStatus.CONFLICT, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("User already exists", response.getBody().message());
		assertEquals("USER_ALREADY_EXISTS", response.getBody().error());
	}

	@Test
	void testHandleUserNotAuthorizedException() {
		UserNotAuthorizedException ex = new UserNotAuthorizedException("Unauthorized");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleUserNotAuthorized(ex, request);

		assertEquals(HttpStatus.FORBIDDEN, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("Unauthorized", response.getBody().message());
		assertEquals("UNAUTHORIZED", response.getBody().error());
	}

	@Test
	void testHandleBadCredentialsException() {
		BadCredentialsException ex = new BadCredentialsException("Bad credentials");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleBadCredentials(ex, request);

		assertEquals(HttpStatus.UNAUTHORIZED, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("Invalid username or password", response.getBody().message());
		assertEquals("BAD_CREDENTIALS", response.getBody().error());
	}

	@Test
	void testHandleAccessDeniedException() {
		AccessDeniedException ex = new AccessDeniedException("Access denied");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleAccessDenied(ex, request);

		assertEquals(HttpStatus.FORBIDDEN, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("Access denied", response.getBody().message());
		assertEquals("ACCESS_DENIED", response.getBody().error());
	}

	@Test
	void testHandleValidationExceptions() {
		FieldError fieldError = new FieldError("user", "name", "invalid value", false, null, null, "Name is required");
		when(validationException.getBindingResult()).thenReturn(bindingResult);
		when(bindingResult.getFieldErrors()).thenReturn(Collections.singletonList(fieldError));

		ResponseEntity<ErrorResponse> response =
				exceptionHandler.handleValidationExceptions(validationException, request);

		assertEquals(HttpStatus.BAD_REQUEST, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("Validation failed", response.getBody().message());
		assertEquals("VALIDATION_ERROR", response.getBody().error());
		assertNotNull(response.getBody().fieldErrors());
		assertEquals(1, response.getBody().fieldErrors().size());
		assertEquals("name", response.getBody().fieldErrors().get(0).field());
		assertEquals("Name is required", response.getBody().fieldErrors().get(0).message());
	}

	@Test
	void testHandleTypeMismatch() {
		MethodArgumentTypeMismatchException ex = mock(MethodArgumentTypeMismatchException.class);
		when(ex.getName()).thenReturn("id");
		when(ex.getValue()).thenReturn("invalid");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleTypeMismatch(ex, request);

		assertEquals(HttpStatus.BAD_REQUEST, response.getStatusCode());
		assertNotNull(response.getBody());
		assertTrue(response.getBody().message().contains("Invalid value 'invalid' for parameter 'id'"));
		assertEquals("TYPE_MISMATCH", response.getBody().error());
	}

	@Test
	void testHandleGenericException() {
		Exception ex = new Exception("Generic error");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleGenericException(ex, request);

		assertEquals(HttpStatus.INTERNAL_SERVER_ERROR, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("An unexpected error occurred", response.getBody().message());
		assertEquals("INTERNAL_SERVER_ERROR", response.getBody().error());
	}

	@Test
	void testHandleIllegalArgumentException() {
		IllegalArgumentException ex = new IllegalArgumentException("Invalid argument");

		ResponseEntity<ErrorResponse> response = exceptionHandler.handleIllegalArgument(ex, request);

		assertEquals(HttpStatus.BAD_REQUEST, response.getStatusCode());
		assertNotNull(response.getBody());
		assertEquals("Invalid argument", response.getBody().message());
		assertEquals("ILLEGAL_ARGUMENT", response.getBody().error());
	}
}
