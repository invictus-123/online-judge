package com.online.judge.backend.repository.specification;

import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.SubmissionLanguage;
import com.online.judge.backend.model.shared.SubmissionStatus;
import java.util.List;
import org.springframework.data.jpa.domain.Specification;

/**
 * Specifications for building dynamic queries for the Submission entity.
 */
public class SubmissionSpecifications {

	/**
	 * Creates a specification that filters submissions by user.
	 * Only submissions by the specified user will be returned.
	 *
	 * @param user The user to filter by
	 * @return Specification for user filtering
	 */
	public static Specification<Submission> hasUser(User user) {
		return (root, query, criteriaBuilder) -> {
			if (user == null) {
				return criteriaBuilder.conjunction();
			}
			return criteriaBuilder.equal(root.get("user"), user);
		};
	}

	/**
	 * Creates a specification that filters submissions by problem ID.
	 * Only submissions for the specified problem will be returned.
	 *
	 * @param problemId The problem ID to filter by
	 * @return Specification for problem filtering
	 */
	public static Specification<Submission> hasProblem(Long problemId) {
		return (root, query, criteriaBuilder) -> {
			if (problemId == null) {
				return criteriaBuilder.conjunction();
			}
			return criteriaBuilder.equal(root.get("problem").get("id"), problemId);
		};
	}

	/**
	 * Creates a specification that filters submissions by status.
	 * Submissions matching any of the provided statuses will be returned.
	 *
	 * @param statuses List of statuses to filter by
	 * @return Specification for status filtering
	 */
	public static Specification<Submission> hasStatusIn(List<SubmissionStatus> statuses) {
		return (root, query, criteriaBuilder) -> {
			if (statuses == null || statuses.isEmpty()) {
				return criteriaBuilder.conjunction();
			}
			return root.get("status").in(statuses);
		};
	}

	/**
	 * Creates a specification that filters submissions by programming language.
	 * Submissions matching any of the provided languages will be returned.
	 *
	 * @param languages List of languages to filter by
	 * @return Specification for language filtering
	 */
	public static Specification<Submission> hasLanguageIn(List<SubmissionLanguage> languages) {
		return (root, query, criteriaBuilder) -> {
			if (languages == null || languages.isEmpty()) {
				return criteriaBuilder.conjunction();
			}
			return root.get("language").in(languages);
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
	public static Specification<Submission> and(Specification<Submission>... specifications) {
		Specification<Submission> result = null;
		for (Specification<Submission> spec : specifications) {
			if (result == null) {
				result = spec;
			} else {
				result = result.and(spec);
			}
		}
		return result != null ? result : (root, query, criteriaBuilder) -> criteriaBuilder.conjunction();
	}

	private SubmissionSpecifications() {
		// Prevent instantiation
	}
}
