package com.online.judge.backend.repository.attributes;

/**
 * Constants for User entity attribute names used in JPA Specifications.
 * Prevents hardcoded strings and enables compile-time checking.
 */
public final class UserAttributes {
	public static final String ID = "id";
	public static final String HANDLE = "handle";
	public static final String EMAIL = "email";

	private UserAttributes() {
		// Prevent instantiation
	}
}
