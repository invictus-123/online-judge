package com.online.judge.backend.dto.filter;

import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import jakarta.validation.constraints.Min;
import java.util.List;

/**
 * Request DTO for filtering problems. This class provides different criterias to filter out
 * problems.
 */
public record ProblemFilterRequest(
		/**
		 * List of difficulties to filter by. Problems matching at least one difficulty
		 * will be returned. If empty or null, no difficulty filtering is applied.
		 */
		List<ProblemDifficulty> difficulties,

		/**
		 * List of tags to filter by. Problems having at least one of these tags
		 * will be returned. If empty or null, no tag filtering is applied.
		 */
		List<ProblemTag> tags,

		/**
		 * The page number to retrieve (1-based).
		 */
		@Min(value = 1, message = "Page number must be at least 1") int page) {

	public ProblemFilterRequest(List<ProblemDifficulty> difficulties, List<ProblemTag> tags) {
		this(difficulties, tags, 1);
	}

	/**
	 * Checks if any difficulty filters are applied.
	 */
	public boolean hasDifficultyFilters() {
		return difficulties != null && !difficulties.isEmpty();
	}

	/**
	 * Checks if any tag filters are applied.
	 */
	public boolean hasTagFilters() {
		return tags != null && !tags.isEmpty();
	}

	/**
	 * Checks if any filters are applied.
	 */
	public boolean hasAnyFilters() {
		return hasDifficultyFilters() || hasTagFilters();
	}
}
