package com.online.judge.backend.model;

import com.online.judge.backend.model.shared.ProblemDifficulty;
import jakarta.persistence.CascadeType;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.FetchType;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Index;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.OneToMany;
import jakarta.persistence.Table;
import jakarta.validation.constraints.Max;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PastOrPresent;
import jakarta.validation.constraints.Size;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.ToString;
import org.hibernate.annotations.CreationTimestamp;

/**
 * Represents the 'problems' table in the database. This class is a JPA entity
 * that maps to problem data.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(
		name = "problems",
		indexes = {
			@Index(name = "idx_problem_created_at", columnList = "createdAt"),
			@Index(name = "idx_problem_difficulty_created_at", columnList = "difficulty,createdAt"),
		})
public class Problem {

	/** The primary key for the problems table. */
	@Id
	@GeneratedValue(strategy = GenerationType.IDENTITY)
	private Long id;

	/** The title of the problem. */
	@NotBlank
	@Size(max = 50)
	@Column(name = "title", nullable = false)
	private String title;

	/** The full problem description, which can include Markdown. */
	@NotBlank
	@Size(max = 50000)
	@Column(name = "statement", nullable = false, columnDefinition = "TEXT")
	private String statement;

	/** The maximum execution time allowed for a solution, in seconds. */
	@Min(1)
	@Max(10)
	@Column(name = "time_limit_second", nullable = false, precision = 2)
	private Double timeLimitSecond;

	/** The maximum memory allowed for a solution, in megabytes. */
	@Min(1)
	@Max(1024)
	@Column(name = "memory_limit_mb", nullable = false)
	private Integer memoryLimitMb;

	/**
	 * The difficulty level of the problem. Stored as a String in the database
	 * (e.g., "EASY", "MEDIUM").
	 */
	@NotNull
	@Enumerated(EnumType.STRING)
	@Column(nullable = false)
	private ProblemDifficulty difficulty;

	/**
	 * The user (admin) who created this problem. This establishes a many-to-one
	 * relationship with the User entity.
	 */
	@ManyToOne(fetch = FetchType.LAZY)
	@JoinColumn(name = "user_id", nullable = false)
	@ToString.Exclude
	private User createdBy;

	/** The timestamp of when the problem was created. */
	@PastOrPresent
	@CreationTimestamp
	@Column(name = "created_at", nullable = false, updatable = false)
	private Instant createdAt;

	/**
	 * The tags associated with this problem. This establishes a one-to-many
	 * relationship with the Tag entity. 'mappedBy = "problem"' indicates that the
	 * Tag entity owns the relationship. 'cascade = CascadeType.ALL' means
	 * operations (like save, delete) on a Problem will cascade to its associated
	 * Tags.
	 */
	@OneToMany(mappedBy = "problem", cascade = CascadeType.ALL, orphanRemoval = true)
	@ToString.Exclude
	private List<Tag> tags = new ArrayList<>();

	/**
	 * The test cases associated with this problem. This establishes a one-to-many
	 * relationship with the TestCase entity. 'mappedBy = "problem"' indicates that
	 * the TestCase entity owns the relationship. 'cascade = CascadeType.ALL' means
	 * operations (like save, delete) on a Problem will cascade to its associated
	 * TestCases.
	 */
	@OneToMany(mappedBy = "problem", cascade = CascadeType.ALL, orphanRemoval = true)
	@ToString.Exclude
	private List<TestCase> testCases = new ArrayList<>();
}
