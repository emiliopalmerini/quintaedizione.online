# Due Draghi - D&D 5e SRD Management System

A comprehensive web-based system for managing and displaying Dungeons & Dragons 5th Edition System Reference Document (SRD) content in Italian. Built with Go, Love and Tears
featuring a clean architecture and modern web technologies.

## 🚀 Features

- **Only Italian Language Support**: Complete Italian SRD content
- **Web Editor**: User-friendly interface for viewing and editing D&D content
- **Content Parser**: Automated processing of markdown files into structured database entries
- **Search & Navigation**: Fast content discovery and browsing
- **Template-Based Rendering**: Clean, responsive web interface with HTMX + Templ
- **Docker Integration**: Containerized deployment with Docker Compose

## 📋 Requirements

- Docker and Docker Compose
- Go 1.24+ (for local development)
- Make (for build commands)

## 🔧 Quick Start

1. **Clone the repository**:
   ```bash
   git clone https://github.com/emiliopalmerini/due-draghi-5e-srd.git
   cd due-draghi-5e-srd
   ```

2. **Initialize environment**:
   ```bash
   make env-init
   ```

3. **Start the services**:
   ```bash
   make up
   ```

4. **Access the applications**:
   - Editor: http://localhost:8000
   - Parser: http://localhost:8100

## 🏗️ Architecture

### Clean Architecture Pattern
```
├── cmd/                    # Application entry points
│   ├── editor/            # Web editor service
│   └── parser/            # Content parser service
├── internal/
│   ├── domain/            # Core business logic and entities
│   ├── application/       # Use cases, services, handlers
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── parsers/       # Content parsing logic
│   │   └── services/      # Business services
│   ├── adapters/          # External interfaces
│   │   └── web/           # Web handlers and routing
│   ├── infrastructure/    # Configuration and setup
│   └── shared/            # Common models and utilities
├── pkg/                   # Reusable packages
│   ├── mongodb/           # MongoDB client and configuration
│   └── templates/         # Template utilities
├── data/                  # SRD content files
│   ├── eng/              # English SRD 5.2 markdown files
│   └── ita/              # Italian SRD markdown files
└── web/                   # Web assets
    ├── static/           # CSS, JS, images
    └── templates/        # HTML templates
```

### Services

#### 1. Editor Service (Port 8000)
- It's really a viewer, but the name is stuck
- Web interface for viewing D&D content
- Template-based rendering with modern web technologies
- User-friendly navigation and search capabilities

#### 2. Parser Service (Port 8100)
- Processes markdown files from the `data/` directory
- Converts content into structured MongoDB documents
- Supports multiple content types: classes, backgrounds, spells, items, etc.

### Database Structure

**MongoDB Collections:**

```
  'cavalcature_e_veicoli',
  'equipaggiamento',
  'documenti',
  'specie',
  'mostri',
  'armature',
  'backgrounds',
  'servizi',
  'incantesimi',
  'animali',
  'armi',
  'talenti',
  'strumenti',
  'oggetti_magici',
  'classi'
]
```

**Document Schema:**
- **BaseEntity**: Common fields (ID, timestamps, version, source)
- **MarkdownContent**: Multi-format content (raw markdown, HTML, plain text)
- **SearchableContent**: Optimized search metadata

## 🛠️ Development

### Available Commands

**Docker Services:**
```bash
make up          # Start MongoDB + Editor + Parser
make down        # Stop all services
make logs        # View service logs
make build       # Build Go images
```

**Go Development:**
```bash
make lint        # Run linting (go vet + golangci-lint)
make test        # Run unit tests
make test-integration  # Run integration tests
make benchmark   # Run performance benchmarks
```

**Database Management:**
```bash
make seed-dump                    # Create timestamped backup
make seed-restore FILE=backup.gz  # Restore from backup
make mongo-sh                     # Access MongoDB shell
```

**Container Access:**
```bash
make editor-sh   # Access editor container
make mongo-sh    # Access MongoDB container
```

### Local Development Setup

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Set up environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run services locally**:
   ```bash
   # Start MongoDB
   docker compose up -d mongo
   
   # Run editor locally
   cd cmd/editor && go run main.go
   
   # Run parser locally  
   cd cmd/parser && go run main.go
   ```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `DB_NAME` | `dnd` | Database name |
| `PORT` | `8000/8100` | Service port |
| `ENVIRONMENT` | `development` | Runtime environment |

## 📝 Content Management

### Data Structure

The system processes D&D content from markdown files organized by language:

**English Content (`data/eng/`):**
- Legal information
- Game rules and mechanics
- Character creation
- Classes and character options
- Spells, equipment, and magic items
- Monsters and creatures

**Italian Content (`data/ita/`):**
- Complete Italian translation of SRD content
- Localized terminology and rules
- Cultural adaptations where appropriate

### Content Processing

The parser service automatically:
1. Reads markdown files from the data directory
2. Extracts structured information (titles, sections, metadata)
3. Converts content to multiple formats (markdown, HTML, plain text)
4. Stores in MongoDB with proper indexing for search
5. Maintains version control and source tracking

## 🧪 Testing

### Unit Tests
```bash
make test
```

### Integration Tests
```bash
make test-integration
```

### Performance Benchmarks
```bash
make benchmark
```

### Manual Testing
1. Start services: `make up`
2. Verify editor at http://localhost:8000
3. Verify parser at http://localhost:8100
4. Test content parsing and display

## 🚀 Deployment

### Docker Production Deployment

1. **Build production images**:
   ```bash
   make build
   ```

2. **Start in production mode**:
   ```bash
   ENVIRONMENT=production make up
   ```

3. **Initialize database** (first time):
   ```bash
   # Access parser to trigger initial content processing
   curl http://localhost:8100/health
   ```

### Database Backup/Restore

**Create Backup:**
```bash
make seed-dump
# Creates timestamped backup: dnd_backup_YYYYMMDD_HHMMSS.archive.gz
```

**Restore Backup:**
```bash
make seed-restore FILE=dnd_backup_20240904_120000.archive.gz
```

## 🔍 Monitoring and Debugging

### View Logs
```bash
make logs                    # All services
docker compose logs editor  # Specific service
docker compose logs -f parser  # Follow logs
```

### Database Access
```bash
make mongo-sh
# In MongoDB shell:
use dnd
db.documenti.find().limit(5)
db.classi.countDocuments()
```

### Health Checks
- Editor health: http://localhost:8000/health
- Parser health: http://localhost:8100/health

## 📄 License

This project contains D&D 5e SRD content licensed under the Creative Commons Attribution 4.0 International License (CC-BY-4.0). See the legal information sections in the content files for complete licensing details.

The application code is available under the terms specified in the project license.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes following the existing code style
4. Run tests: `make test`
5. Run linting: `make lint`
6. Commit your changes: `git commit -m "feat: your feature description"`
7. Push to your fork: `git push origin feature/your-feature`
8. Create a pull request

### Code Style
- Follow Go best practices and conventions
- Use clean architecture principles
- Write comprehensive tests for new features
- Update documentation for significant changes

## 📚 Additional Resources

- [D&D 5e SRD Official](https://dnd.wizards.com/resources/systems-reference-document)
- [Go Documentation](https://golang.org/doc/)
- [MongoDB Documentation](https://docs.mongodb.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

## 🐛 Troubleshooting

### Common Issues

**Services won't start:**
```bash
# Check if ports are in use
netstat -tulpn | grep :8000
netstat -tulpn | grep :8100

# Reset Docker environment
make down
docker system prune -f
make up
```

**MongoDB connection issues:**
```bash
# Verify MongoDB is running
docker compose ps mongo

# Check MongoDB logs
docker compose logs mongo

# Test connection
make mongo-sh
```

**Content not appearing:**
```bash
# Trigger parser manually
curl -X POST http://localhost:8100/parse

# Check database content
make mongo-sh
# In MongoDB shell: db.documenti.countDocuments()
```

For additional support, please open an issue in the GitHub repository.
