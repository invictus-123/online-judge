package com.online.judge.backend.service;

import static com.online.judge.backend.converter.SubmissionConverter.toSubmissionDetailsUi;
import static com.online.judge.backend.converter.SubmissionConverter.toSubmissionFromRequest;
import static com.online.judge.backend.converter.SubmissionConverter.toSubmissionMessage;

import com.online.judge.backend.converter.SubmissionConverter;
import com.online.judge.backend.dto.request.SubmitCodeRequest;
import com.online.judge.backend.dto.ui.SubmissionDetailsUi;
import com.online.judge.backend.dto.ui.SubmissionSummaryUi;
import com.online.judge.backend.exception.ProblemNotFoundException;
import com.online.judge.backend.exception.SubmissionNotFoundException;
import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.SubmissionStatus;
import com.online.judge.backend.queue.SubmissionPublisher;
import com.online.judge.backend.repository.ProblemRepository;
import com.online.judge.backend.repository.SubmissionRepository;
import com.online.judge.backend.util.UserUtil;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** Service class for managing submissions. */
@Service
public class SubmissionService {
	private static final Logger logger = LoggerFactory.getLogger(SubmissionService.class);

	private final ProblemRepository problemRepository;
	private final SubmissionRepository submissionRepository;
	private final SubmissionPublisher submissionPublisher;
	private final UserUtil userUtil;
	private final int pageSize;

	public SubmissionService(
			ProblemRepository problemRepository,
			SubmissionRepository submissionRepository,
			SubmissionPublisher submissionPublisher,
			UserUtil userUtil,
			@Value("${submissions.list.page-size:50}") int pageSize) {
		this.problemRepository = problemRepository;
		this.submissionRepository = submissionRepository;
		this.submissionPublisher = submissionPublisher;
		this.userUtil = userUtil;
		this.pageSize = pageSize;
	}

	/**
	 * Retrieves a paginated list of all submissions, sorted by the submission time in descending
	 * order.
	 *
	 * @param page
	 *            The page number (1-based).
	 * @return A List<SubmissionSummaryUi> containing the paginated list of submissions.
	 */
	@Transactional(readOnly = true)
	public List<SubmissionSummaryUi> listSubmissions(int page) {
		logger.info("Fetching all submissions for page {}", page);

		Pageable pageable =
				PageRequest.of(page - 1, pageSize, Sort.by("submittedAt").descending());
		return submissionRepository.findAll(pageable).getContent().stream()
				.map(SubmissionConverter::toSubmissionSummaryUi)
				.toList();
	}

	/**
	 * Retrieves the details of a submission by its ID.
	 *
	 * @param submissionId
	 *            The ID of the submission to retrieve.
	 * @return A SubmissionDetailsUi object containing the submission details.
	 * @throws SubmissionNotFoundException if the submission with the given ID does not exist.
	 */
	@Transactional(readOnly = true)
	public SubmissionDetailsUi getSubmissionDetailsById(Long submissionId) {
		logger.info("Fetching details for submission with ID: {}", submissionId);

		return submissionRepository
				.findById(submissionId)
				.map(submission -> {
					logger.info("Submission found: {}", submission);
					return toSubmissionDetailsUi(submission);
				})
				.orElseThrow(() -> toSubmissionNotFoundException(submissionId));
	}

	/**
	 * Creates a new submission for a problem.
	 *
	 * @param request
	 *               The request containing the code, language, and problem ID.
	 * @return A DTO representing the newly created submission's details.
	 */
	@Transactional
	public SubmissionDetailsUi submitCode(SubmitCodeRequest request) {
		User currentUser = userUtil.getCurrentAuthenticatedUser();
		Problem problem = problemRepository.findById(request.problemId()).orElseThrow(() -> {
			logger.error("Could not submit code, problem with ID {} not found", request.problemId());
			return new ProblemNotFoundException("Problem with ID " + request.problemId() + " not found");
		});

		Submission submission = toSubmissionFromRequest(request);
		submission.setProblem(problem);
		submission.setUser(currentUser);
		Submission savedSubmission = submissionRepository.save(submission);
		logger.info(
				"Code submitted with ID: {} for problem: {} by user: {}",
				savedSubmission.getId(),
				problem.getId(),
				currentUser.getHandle());

		enqueueSubmission(savedSubmission);

		return toSubmissionDetailsUi(savedSubmission);
	}

	/**
	 * Updates the status for a particular submission.
	 *
	 * @param submissionId
	 *             The ID of the submission
	 * @param staus
	 *             The status of the submission
	 */
	@Transactional
	public void updateStatus(Long submissionId, SubmissionStatus status) {
		Submission submission = submissionRepository
				.findById(submissionId)
				.orElseThrow(() -> toSubmissionNotFoundException(submissionId));

		logger.info("Updating status for submission {} from {} to {}.", submissionId, submission.getStatus(), status);
		submission.setStatus(status);
		submissionRepository.save(submission);
	}

	/**
	 * Updates the time taken and memory used by the overall submission.
	 *
	 * @param submissionId
	 *             The ID of the submission
	 * @param timeTaken
	 *             The time taken for the entire submission (cummulative of all test cases)
	 * @param memoryUsed
	 *             The memory for the entire submission (cummulative of all test cases)
	 */
	@Transactional
	public void updateTimeTakenAndMemoryUsed(Long submissionId, Double timeTaken, Integer memoryUsed) {
		Submission submission = submissionRepository
				.findById(submissionId)
				.orElseThrow(() -> toSubmissionNotFoundException(submissionId));

		logger.info(
				"Update time taken and memory used for submission {} to {} and {} respectively.",
				submissionId,
				timeTaken,
				memoryUsed);
		submission.setExecutionTimeSeconds(timeTaken);
		submission.setMemoryUsedMb(memoryUsed);
		submissionRepository.save(submission);
	}

	private static SubmissionNotFoundException toSubmissionNotFoundException(Long submissionId) {
		logger.error("Submission with ID {} not found", submissionId);
		return new SubmissionNotFoundException("Submission with ID " + submissionId + " not found");
	}

	private void enqueueSubmission(Submission submission) {
		submissionPublisher.sendSubmission(toSubmissionMessage(submission));
	}
}
