package com.online.judge.backend.dto.response;

import com.fasterxml.jackson.annotation.JsonInclude;
import java.time.LocalDateTime;
import java.util.List;

@JsonInclude(JsonInclude.Include.NON_NULL)
public record ErrorResponse(
		String message, String error, int status, LocalDateTime timestamp, String path, List<FieldError> fieldErrors) {

	public ErrorResponse(String message, String error, int status, String path) {
		this(message, error, status, LocalDateTime.now(), path, null);
	}

	public ErrorResponse(String message, String error, int status, String path, List<FieldError> fieldErrors) {
		this(message, error, status, LocalDateTime.now(), path, fieldErrors);
	}

	public record FieldError(String field, Object rejectedValue, String message) {}
}
