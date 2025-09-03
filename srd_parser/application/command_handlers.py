"""
Command handlers for Parser service following hexagonal architecture
Implements the application layer between domain and infrastructure
"""
from typing import List, Optional
import logging
from dataclasses import dataclass

from shared_domain.entities import DndClass, ClassRepository, EventPublisher
from shared_domain.use_cases import ParseClassCommand, ParseClassUseCase, UseCaseResult
from srd_parser.domain.services import ClassParsingService

logger = logging.getLogger(__name__)


@dataclass
class ParseMultipleClassesCommand:
    """Command to parse multiple classes from markdown content"""
    markdown_lines: List[str]
    source: str = "SRD"
    dry_run: bool = False


class ParseMultipleClassesHandler:
    """Handler for parsing multiple classes from markdown"""
    
    def __init__(
        self,
        class_repository: ClassRepository,
        event_publisher: EventPublisher,
        parsing_service: ClassParsingService
    ):
        self.class_repository = class_repository
        self.event_publisher = event_publisher
        self.parsing_service = parsing_service
    
    async def handle(self, command: ParseMultipleClassesCommand) -> UseCaseResult:
        """Parse and save multiple classes from markdown content"""
        try:
            logger.info(f"Starting batch parse of classes (dry_run: {command.dry_run})")
            
            # Parse classes using domain service
            parsed_classes = self.parsing_service.parse_classes_from_markdown(
                command.markdown_lines,
                source=command.source
            )
            
            logger.info(f"Successfully parsed {len(parsed_classes)} classes")
            
            # Save each class using individual use case
            parse_use_case = ParseClassUseCase(self.class_repository, self.event_publisher)
            results = []
            
            for dnd_class in parsed_classes:
                class_command = ParseClassCommand(
                    class_data=dnd_class,
                    dry_run=command.dry_run
                )
                result = await parse_use_case.execute(class_command)
                results.append(result)
                
                if not result.success:
                    logger.warning(f"Failed to process class {dnd_class.name}: {result.error}")
            
            successful_results = [r for r in results if r.success]
            failed_results = [r for r in results if not r.success]
            
            return UseCaseResult(
                success=len(failed_results) == 0,
                data={
                    "total_parsed": len(parsed_classes),
                    "successful": len(successful_results),
                    "failed": len(failed_results),
                    "classes": [r.data for r in successful_results if r.data]
                },
                error=f"{len(failed_results)} classes failed to process" if failed_results else None
            )
            
        except Exception as e:
            logger.error(f"Error in batch class parsing: {e}", exc_info=True)
            return UseCaseResult(
                success=False,
                error=f"Batch parsing failed: {str(e)}"
            )


@dataclass
class ValidateClassDataCommand:
    """Command to validate class data without saving"""
    markdown_lines: List[str]
    source: str = "SRD"


class ValidateClassDataHandler:
    """Handler for validating class data without persistence"""
    
    def __init__(self, parsing_service: ClassParsingService):
        self.parsing_service = parsing_service
    
    async def handle(self, command: ValidateClassDataCommand) -> UseCaseResult:
        """Validate class data and return validation results"""
        try:
            logger.info("Starting class data validation")
            
            # Parse classes for validation
            parsed_classes = self.parsing_service.parse_classes_from_markdown(
                command.markdown_lines,
                source=command.source
            )
            
            validation_results = []
            for dnd_class in parsed_classes:
                # Validate using domain rules
                validation_result = {
                    "class_name": dnd_class.name,
                    "valid": True,
                    "warnings": [],
                    "errors": []
                }
                
                # Basic validation checks
                if not dnd_class.name.strip():
                    validation_result["errors"].append("Class name is empty")
                    validation_result["valid"] = False
                
                if dnd_class.hit_die.value < 6 or dnd_class.hit_die.value > 12:
                    validation_result["warnings"].append(f"Unusual hit die: {dnd_class.hit_die.value}")
                
                if not dnd_class.features:
                    validation_result["warnings"].append("No class features found")
                
                if len(dnd_class.features) < 5:
                    validation_result["warnings"].append("Very few class features found")
                
                validation_results.append(validation_result)
            
            # Summary statistics
            total_classes = len(validation_results)
            valid_classes = len([r for r in validation_results if r["valid"]])
            classes_with_warnings = len([r for r in validation_results if r["warnings"]])
            
            return UseCaseResult(
                success=True,
                data={
                    "total_classes": total_classes,
                    "valid_classes": valid_classes,
                    "classes_with_warnings": classes_with_warnings,
                    "validation_results": validation_results
                }
            )
            
        except Exception as e:
            logger.error(f"Error in class data validation: {e}", exc_info=True)
            return UseCaseResult(
                success=False,
                error=f"Validation failed: {str(e)}"
            )