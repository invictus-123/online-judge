package com.online.judge.backend.repository.attributes;

/**
 * Constants for Tag entity attribute names used in JPA Specifications.
 * Prevents hardcoded strings and enables compile-time checking.
 */
public final class TagAttributes {
	public static final String ID = "id";
	public static final String TAG_NAME = "tagName";

	private TagAttributes() {
		// Prevent instantiation
	}
}
