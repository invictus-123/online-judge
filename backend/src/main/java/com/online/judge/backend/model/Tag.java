package com.online.judge.backend.model;

import com.online.judge.backend.model.shared.ProblemTag;
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
import jakarta.validation.constraints.NotNull;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Represents the 'tags' table in the database. This class is a JPA entity that
 * maps to tag data.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(
		name = "tags",
		indexes = {
			@Index(name = "idx_tag_name", columnList = "tag_name"),
			@Index(name = "idx_tag_problem_tag", columnList = "problem_id,tag_name")
		})
public class Tag {

	/** The primary key for the tags table. */
	@Id
	@GeneratedValue(strategy = GenerationType.UUID)
	private UUID id;

	/**
	 * The problem associated with this tag. Each problem can have multiple tags.
	 */
	@ManyToOne(fetch = FetchType.LAZY)
	@JoinColumn(name = "problem_id", nullable = false)
	private Problem problem;

	/** The name of the tag, represented as an enum. */
	@NotNull
	@Enumerated(EnumType.STRING)
	@Column(name = "tag_name", nullable = false)
	private ProblemTag tagName;
}
