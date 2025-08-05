package com.online.judge.backend.dto.filter;

import com.online.judge.backend.model.shared.SubmissionLanguage;
import com.online.judge.backend.model.shared.SubmissionStatus;
import jakarta.validation.constraints.Min;
import java.util.List;

/**
 * DTO for filtering submissions with optional criteria. This class provides different criterias to 
 * filter out submissions. All filter parameters are optional - null or empty lists are treated
 * as "no filter applied".
 */
public record SubmissionFilterRequest(
		/**
		 * Filter to show only submissions by the current user.
		 * If true, only submissions by the authenticated user are returned.
		 * If false or null, submissions by all users are returned.
		 * If no user is authenticated and this filter is true, it will be ignored.
		 */
		Boolean onlyMe,

		/**
		 * Filter submissions by problem ID.
		 * If specified, only submissions for this problem are returned.
		 */
		Long problemId,

		/**
		 * Filter submissions by status.
		 * Submissions matching any of the provided statuses will be returned.
		 * If null or empty, no status filtering is applied.
		 */
		List<SubmissionStatus> statuses,

		/**
		 * Filter submissions by programming language.
		 * Submissions matching any of the provided languages will be returned.
		 * If null or empty, no language filtering is applied.
		 */
		List<SubmissionLanguage> languages,

		/**
		 * The page number for pagination (1-based).
		 * Must be at least 1.
		 */
		@Min(value = 1, message = "Page number must be at least 1") int page) {}
