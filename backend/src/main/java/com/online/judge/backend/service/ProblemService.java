package com.online.judge.backend.service;

import static com.online.judge.backend.converter.ProblemConverter.toProblemDetailsUi;
import static com.online.judge.backend.converter.ProblemConverter.toProblemFromCreateProblemRequest;
import static com.online.judge.backend.repository.specification.ProblemSpecifications.and;
import static com.online.judge.backend.repository.specification.ProblemSpecifications.hasDifficultyIn;
import static com.online.judge.backend.repository.specification.ProblemSpecifications.hasTagIn;

import com.online.judge.backend.converter.ProblemConverter;
import com.online.judge.backend.dto.filter.ProblemFilterRequest;
import com.online.judge.backend.dto.request.CreateProblemRequest;
import com.online.judge.backend.dto.ui.ProblemDetailsUi;
import com.online.judge.backend.dto.ui.ProblemSummaryUi;
import com.online.judge.backend.exception.ProblemNotFoundException;
import com.online.judge.backend.exception.UserNotAuthorizedException;
import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.UserRole;
import com.online.judge.backend.repository.ProblemRepository;
import com.online.judge.backend.util.UserUtil;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** Service class for managing problems. */
@Service
public class ProblemService {
	private static final Logger logger = LoggerFactory.getLogger(ProblemService.class);

	private final ProblemRepository problemRepository;
	private final UserUtil userUtil;
	private final int pageSize;

	public ProblemService(
			ProblemRepository problemRepository,
			UserUtil userUtil,
			@Value("${problems.list.page-size:50}") int pageSize) {
		this.problemRepository = problemRepository;
		this.userUtil = userUtil;
		this.pageSize = pageSize;
	}

	/**
	 * Retrieves a paginated list of problems with optional filtering, sorted by creation date in
	 * descending order.
	 *
	 * @param filterRequest The filter request containing pagination and filter criteria.
	 * @return A List<ProblemSummaryUi> containing the paginated list of problems matching the filters.
	 */
	@Transactional(readOnly = true)
	public List<ProblemSummaryUi> listProblems(ProblemFilterRequest filterRequest) {
		logger.info(
				"Fetching problems with filters: page={}, difficulties={}, tags={}",
				filterRequest.page(),
				filterRequest.difficulties(),
				filterRequest.tags());

		Pageable pageable = PageRequest.of(
				filterRequest.page() - 1, pageSize, Sort.by("createdAt").descending());

		// Build the specification dynamically based on filter criteria
		Specification<Problem> specification =
				and(hasDifficultyIn(filterRequest.difficulties()), hasTagIn(filterRequest.tags()));

		return problemRepository.findAll(specification, pageable).getContent().stream()
				.map(ProblemConverter::toProblemSummaryUi)
				.toList();
	}

	/**
	 * Retrieves the details of a problem by its ID.
	 *
	 * @param problemId
	 *            The ID of the problem to retrieve.
	 * @return A ProblemDetailsUi object containing the problem details.
	 * @throws ProblemNotFoundException
	 *             if the problem with the given ID does not exist.
	 */
	@Transactional(readOnly = true)
	public ProblemDetailsUi getProblemDetailsById(Long problemId) {
		logger.info("Fetching details for problem with ID: {}", problemId);

		return problemRepository
				.findById(problemId)
				.map(problem -> {
					logger.info("Problem found: {}", problem);
					return toProblemDetailsUi(problem);
				})
				.orElseThrow(() -> {
					logger.error("Problem with ID {} not found", problemId);
					return new ProblemNotFoundException("Problem with ID " + problemId + " not found");
				});
	}

	/**
	 * Creates a new problem, including its tags and test cases. This method is transactional to
	 * ensure all data is saved atomically.
	 *
	 * @param request
	 *            The DTO record containing all information for the new problem.
	 * @return A DTO representing the newly created problem's details.
	 */
	@Transactional
	public ProblemDetailsUi createProblem(CreateProblemRequest request) {
		logger.info("Creating a new problem with title: {}", request.title());

		User authenticatedUser = userUtil.getCurrentAuthenticatedUser();
		if (!authenticatedUser.getRole().equals(UserRole.ADMIN)) {
			logger.warn("User {} is not authorized to create problems", authenticatedUser.getHandle());
			throw new UserNotAuthorizedException("User is not authorized to create problems.");
		}

		Problem problem = toProblemFromCreateProblemRequest(request);
		problem.setCreatedBy(authenticatedUser);
		Problem savedProblem = problemRepository.save(problem);
		logger.info("Problem created with ID: {} by user: {}", savedProblem.getId(), authenticatedUser.getHandle());
		return toProblemDetailsUi(savedProblem);
	}
}
