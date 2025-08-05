package com.online.judge.backend.repository.specification;

import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.Tag;
import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import jakarta.persistence.criteria.Join;
import jakarta.persistence.criteria.JoinType;
import java.util.List;
import org.springframework.data.jpa.domain.Specification;

/**
 * Specifications for building dynamic queries for the Problem entity.
 */
public class ProblemSpecifications {

	/**
	 * Creates a specification that filters problems by difficulty.
	 * Problems matching at least one of the provided difficulties will be returned.
	 *
	 * @param difficulties List of difficulties to filter by
	 * @return Specification for difficulty filtering
	 */
	public static Specification<Problem> hasDifficultyIn(List<ProblemDifficulty> difficulties) {
		return (root, query, criteriaBuilder) -> {
			if (difficulties == null || difficulties.isEmpty()) {
				return criteriaBuilder.conjunction();
			}
			return root.get("difficulty").in(difficulties);
		};
	}

	/**
	 * Creates a specification that filters problems by tags.
	 * Problems having at least one of the provided tags will be returned.
	 *
	 * @param tags List of tags to filter by
	 * @return Specification for tag filtering
	 */
	public static Specification<Problem> hasTagIn(List<ProblemTag> tags) {
		return (root, query, criteriaBuilder) -> {
			if (tags == null || tags.isEmpty()) {
				return criteriaBuilder.conjunction();
			}

			Join<Problem, Tag> tagJoin = root.join("tags", JoinType.INNER);
			query.distinct(true);
			return tagJoin.get("tagName").in(tags);
		};
	}

	/**
	 * Combines multiple specifications using AND logic.
	 * This method provides a convenient way to combine multiple filter specifications.
	 *
	 * @param specifications Variable number of specifications to combine
	 * @return Combined specification
	 */
	@SafeVarargs
	public static Specification<Problem> and(Specification<Problem>... specifications) {
		Specification<Problem> result = null;
		for (Specification<Problem> spec : specifications) {
			if (result == null) {
				result = spec;
			} else {
				result = result.and(spec);
			}
		}
		return result != null ? result : (root, query, criteriaBuilder) -> criteriaBuilder.conjunction();
	}

	private ProblemSpecifications() {
		// Prevent instantiation
	}
}
