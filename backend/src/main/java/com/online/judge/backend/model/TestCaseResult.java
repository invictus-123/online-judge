package com.online.judge.backend.model;

import com.online.judge.backend.model.shared.SubmissionStatus;
import com.online.judge.backend.model.shared.TestCaseResultId;
import jakarta.persistence.Column;
import jakarta.persistence.EmbeddedId;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.FetchType;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.MapsId;
import jakarta.persistence.Table;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Represents the 'test_case_results' table in the database. Each row links a
 * submission and a test case with the result.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "test_case_results")
public class TestCaseResult {

	@EmbeddedId
	private TestCaseResultId id;

	@ManyToOne(fetch = FetchType.LAZY)
	@MapsId("submissionId")
	@JoinColumn(name = "submission_id", nullable = false)
	private Submission submission;

	@ManyToOne(fetch = FetchType.LAZY)
	@MapsId("testCaseId")
	@JoinColumn(name = "test_case_id", nullable = false)
	private TestCase testCase;

	@NotNull
	@Enumerated(EnumType.STRING)
	@Column(name = "verdict", nullable = false)
	private SubmissionStatus verdict;

	@Column(name = "user_output", columnDefinition = "TEXT")
	private String userOutput;

	@Column(name = "checker_log", columnDefinition = "TEXT")
	private String checkerLog;

	@Column(name = "execution_time_seconds", precision = 2)
	private Double executionTimeSeconds;

	@Column(name = "memory_used_mb")
	private Integer memoryUsedMb;
}
