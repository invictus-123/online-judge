package com.online.judge.backend.repository.specification;

import jakarta.persistence.criteria.Join;
import jakarta.persistence.criteria.JoinType;
import java.util.List;
import org.springframework.data.jpa.domain.Specification;

/**
 * Base utility class providing common specification methods for all entities.
 */
public final class BaseSpecifications {

	/**
	 * Creates a specification that checks if an attribute value is in the provided collection.
	 *
	 * @param attributeName The name of the entity attribute to check
	 * @param values List of values to match against
	 * @return Specification for value matching
	 */
	public static <T, V> Specification<T> hasAttributeInValues(String attributeName, List<V> values) {
		return (root, query, criteriaBuilder) -> {
			if (values == null || values.isEmpty()) {
				return criteriaBuilder.conjunction();
			}
			return root.get(attributeName).in(values);
		};
	}

	/**
	 * Creates a specification that checks if an attribute has a specific value.
	 *
	 * @param attributeName The name of the entity attribute to check
	 * @param value The value to match against
	 * @return Specification for exact value matching
	 */
	public static <T, V> Specification<T> hasAttributeWithValue(String attributeName, V value) {
		return (root, query, criteriaBuilder) -> {
			if (value == null) {
				return criteriaBuilder.conjunction();
			}
			return criteriaBuilder.equal(root.get(attributeName), value);
		};
	}

	/**
	 * Creates a specification that checks if a nested attribute has a specific value.
	 *
	 * @param parentAttribute The name of the parent entity attribute
	 * @param childAttribute The name of the child attribute to check
	 * @param value The value to match against
	 * @return Specification for nested attribute matching
	 */
	public static <T, V> Specification<T> hasNestedAttributeWithValue(
			String parentAttribute, String childAttribute, V value) {
		return (root, query, criteriaBuilder) -> {
			if (value == null) {
				return criteriaBuilder.conjunction();
			}
			return criteriaBuilder.equal(root.get(parentAttribute).get(childAttribute), value);
		};
	}

	/**
	 * Creates a specification that checks if a nested attribute value is in the provided collection.
	 *
	 * @param parentAttribute The name of the parent entity attribute
	 * @param childAttribute The name of the child attribute to check
	 * @param values List of values to match against
	 * @return Specification for nested attribute collection matching
	 */
	public static <T, V> Specification<T> hasNestedAttributeInValues(
			String parentAttribute, String childAttribute, List<V> values) {
		return (root, query, criteriaBuilder) -> {
			if (values == null || values.isEmpty()) {
				return criteriaBuilder.conjunction();
			}
			return root.get(parentAttribute).get(childAttribute).in(values);
		};
	}

	/**
	 * Creates a specification that checks if values in a joined entity match the provided
	 * collection.
	 *
	 * @param joinAttribute The attribute to join on
	 * @param targetAttribute The attribute in the joined entity to check
	 * @param values List of values to match against
	 * @param joinType The type of join to perform
	 * @return Specification for joined entity filtering
	 */
	public static <T, V> Specification<T> hasJoinedAttributeInValues(
			String joinAttribute, String targetAttribute, List<V> values, JoinType joinType) {
		return (root, query, criteriaBuilder) -> {
			if (values == null || values.isEmpty()) {
				return criteriaBuilder.conjunction();
			}

			Join<T, ?> join = root.join(joinAttribute, joinType);
			query.distinct(true);
			return join.get(targetAttribute).in(values);
		};
	}

	/**
	 * Combines multiple specifications using AND logic.
	 *
	 * @param specifications Variable number of specifications to combine
	 * @return Combined specification
	 */
	@SafeVarargs
	public static <T> Specification<T> and(Specification<T>... specifications) {
		Specification<T> result = null;
		for (Specification<T> spec : specifications) {
			if (result == null) {
				result = spec;
			} else {
				result = result.and(spec);
			}
		}
		return result != null ? result : (root, query, criteriaBuilder) -> criteriaBuilder.conjunction();
	}

	/**
	 * Combines multiple specifications using OR logic.
	 *
	 * @param specifications Variable number of specifications to combine
	 * @return Combined specification
	 */
	@SafeVarargs
	public static <T> Specification<T> or(Specification<T>... specifications) {
		Specification<T> result = null;
		for (Specification<T> spec : specifications) {
			if (result == null) {
				result = spec;
			} else {
				result = result.or(spec);
			}
		}
		return result != null ? result : (root, query, criteriaBuilder) -> criteriaBuilder.disjunction();
	}

	private BaseSpecifications() {
		// Prevent instantiation
	}
}
