package com.online.judge.backend.service;

import static com.online.judge.backend.converter.TestCaseResultConverter.toTestCaseResult;

import com.online.judge.backend.dto.message.TestCaseResultMessage;
import com.online.judge.backend.exception.SubmissionNotFoundException;
import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.TestCaseResult;
import com.online.judge.backend.repository.SubmissionRepository;
import com.online.judge.backend.repository.TestCaseResultRepository;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** Service class for managing test case results. */
@Service
public class TestCaseResultService {
	private static final Logger logger = LoggerFactory.getLogger(TestCaseResultService.class);

	private final SubmissionRepository submissionRepository;
	private final TestCaseResultRepository testCaseResultRepository;

	public TestCaseResultService(
			SubmissionRepository submissionRepository, TestCaseResultRepository testCaseResultRepository) {
		this.submissionRepository = submissionRepository;
		this.testCaseResultRepository = testCaseResultRepository;
	}

	/**
	 * Updates the status for a particular submission.
	 *
	 * @param submissionId
	 *             ID of the submission
	 * @param testCaseResults
	 *             Result of the individual test case execution
	 */
	@Transactional
	public void processTestResult(Long submissionId, List<TestCaseResultMessage> testCaseResults) {
		Submission submission = submissionRepository
				.findById(submissionId)
				.orElseThrow(() -> toSubmissionNotFoundException(submissionId));

		List<TestCaseResult> testCaseResultEntities = toTestCaseResult(submission, testCaseResults);
		testCaseResultRepository.saveAll(testCaseResultEntities);
	}

	private static SubmissionNotFoundException toSubmissionNotFoundException(Long submissionId) {
		logger.error("Submission with ID {} not found", submissionId);
		return new SubmissionNotFoundException("Submission with ID " + submissionId + " not found");
	}
}
