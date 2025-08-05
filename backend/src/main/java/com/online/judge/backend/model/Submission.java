package com.online.judge.backend.model;

import com.online.judge.backend.model.shared.SubmissionLanguage;
import com.online.judge.backend.model.shared.SubmissionStatus;
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
import jakarta.persistence.Table;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PastOrPresent;
import java.time.Instant;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;

/**
 * Represents the 'submissions' table in the database. Each submission is associated with a user and
 * a problem.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(
		name = "submissions",
		indexes = {
			@Index(name = "idx_submission_submitted_at", columnList = "submittedAt"),
			@Index(name = "idx_submission_user_submitted_at", columnList = "user_id,submittedAt"),
			@Index(name = "idx_submission_problem_submitted_at", columnList = "problem_id,submittedAt"),
			@Index(name = "idx_submission_status_submitted_at", columnList = "status,submittedAt"),
			@Index(name = "idx_submission_language_submitted_at", columnList = "language,submittedAt"),
		})
public class Submission {

	/** The primary key for the submissions table. */
	@Id
	@GeneratedValue(strategy = GenerationType.IDENTITY)
	private Long id;

	/** The user who made the submission. */
	@ManyToOne(fetch = FetchType.LAZY)
	@JoinColumn(name = "user_id", nullable = false)
	private User user;

	/** The problem this submission is for. */
	@ManyToOne(fetch = FetchType.LAZY)
	@JoinColumn(name = "problem_id", nullable = false)
	private Problem problem;

	/** The timestamp when the code was submitted. */
	@PastOrPresent
	@CreationTimestamp
	@Column(name = "submitted_at", nullable = false, updatable = false)
	private Instant submittedAt;

	/** The status of the submission. */
	@NotNull
	@Enumerated(EnumType.STRING)
	@Column(name = "status", nullable = false)
	private SubmissionStatus status;

	/** The programming language used for the submission. */
	@NotNull
	@Enumerated(EnumType.STRING)
	@Column(name = "language", nullable = false)
	private SubmissionLanguage language;

	/** The code submitted by the user. */
	@NotBlank
	@Column(name = "code", columnDefinition = "TEXT", nullable = false)
	private String code;

	/** The execution time in seconds (nullable, only set after execution). */
	@Column(name = "execution_time_seconds")
	private Double executionTimeSeconds;

	/** The memory used in MB (nullable, only set after execution). */
	@Column(name = "memory_used_mb")
	private Integer memoryUsedMb;
}
