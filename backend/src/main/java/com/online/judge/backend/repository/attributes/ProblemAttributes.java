package com.online.judge.backend.repository.attributes;

/**
 * Constants for Problem entity attribute names used in JPA Specifications.
 * Prevents hardcoded strings and enables compile-time checking.
 */
public final class ProblemAttributes {
	public static final String ID = "id";
	public static final String TITLE = "title";
	public static final String DESCRIPTION = "description";
	public static final String DIFFICULTY = "difficulty";
	public static final String TIME_LIMIT_SECONDS = "timeLimitSeconds";
	public static final String MEMORY_LIMIT_MB = "memoryLimitMb";
	public static final String CREATED_AT = "createdAt";
	public static final String CREATED_BY = "createdBy";
	public static final String TAGS = "tags";
	public static final String TEST_CASES = "testCases";
	public static final String SUBMISSIONS = "submissions";

	private ProblemAttributes() {
		// Prevent instantiation
	}
}
