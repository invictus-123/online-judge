package com.online.judge.backend.repository.attributes;

/**
 * Constants for Submission entity attribute names used in JPA Specifications.
 * Prevents hardcoded strings and enables compile-time checking.
 */
public final class SubmissionAttributes {
	public static final String ID = "id";
	public static final String CODE = "code";
	public static final String LANGUAGE = "language";
	public static final String STATUS = "status";
	public static final String EXECUTION_TIME_MS = "executionTimeMs";
	public static final String MEMORY_USAGE_KB = "memoryUsageKb";
	public static final String SUBMITTED_AT = "submittedAt";
	public static final String USER = "user";
	public static final String PROBLEM = "problem";
	public static final String TEST_CASE_RESULTS = "testCaseResults";

	private SubmissionAttributes() {
		// Prevent instantiation
	}
}
