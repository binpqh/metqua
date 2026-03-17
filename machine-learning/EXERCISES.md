# 🎓 Machine Learning Exercises & Projects

## 📚 Exercises by Topic

Mỗi bài tập bao gồm:

- ✅ **Đề bài** với dataset cụ thể
- ✅ **Approach/Solution** chi tiết
- ✅ **Implementation guide** từng bước
- ✅ **Complete code** solution
- ✅ **Extension challenges** để practice thêm

### Danh sách bài tập:

1. **[Data Preprocessing](0-preprocessing/EXERCISE.md)** ✅
   - Xử lý missing values
   - Encode categorical features
   - Feature scaling
   - Train/test split
   - **Dataset**: Student Performance

2. **[Regression](1-regressions/EXERCISE.md)** ✅
   - So sánh 3 models: Linear Regression, Random Forest, SVR
   - Predict giá nhà
   - Visualize predictions
   - **Dataset**: House Prices

3. **[Classification](2-classifications/EXERCISE.md)** ✅
   - Email spam detection
   - Text preprocessing with Bag of Words
   - So sánh Naive Bayes, Logistic Regression, Random Forest
   - Confusion matrix & metrics analysis
   - **Dataset**: Email spam/ham

4. **[Clustering](3-clustering/EXERCISE.md)** ✅
   - Customer segmentation for marketing
   - Elbow Method to find optimal K
   - K-Means clustering & visualization
   - Business insights & strategy
   - **Dataset**: Customer demographics & spending

5. **[Association Rules](4-association-rule-learning/EXERCISE.md)** ✅
   - Market basket analysis
   - Apriori algorithm
   - Support, Confidence, Lift metrics
   - Product bundles & recommendations
   - **Dataset**: Grocery transactions

6. **[Reinforcement Learning](5-reinforcement-learning/EXERCISE.md)** ✅
   - Online ad campaign optimization
   - Multi-Armed Bandit problem
   - Compare Random, UCB, Thompson Sampling
   - Regret analysis & business impact
   - **Dataset**: Simulated ad clicks

7. **[NLP](6-natural-language-processing/EXERCISE.md)** ✅
   - Social media sentiment analysis
   - Text cleaning, stemming, TF-IDF
   - Sentiment classification (Positive/Negative)
   - Feature importance & word clouds
   - **Dataset**: Customer reviews

8. **[Model Selection](9-model-selection/EXERCISE.md)** ✅
   - Credit card fraud detection
   - K-Fold Cross Validation
   - Grid Search hyperparameter tuning
   - Compare Logistic Regression, Random Forest, XGBoost
   - ROC-AUC analysis
   - **Dataset**: Imbalanced fraud transactions

9. **Deep Learning** _(Coming Soon)_
   - Artificial Neural Networks (ANN)
   - Convolutional Neural Networks (CNN)

10. **Dimensionality Reduction** _(Coming Soon)_

- PCA vs LDA comparison
- Feature extraction & visualization

---

## 🚀 End-to-End Production Project

**[Customer Churn Prediction & Deployment](end-to-end-project/)**

Complete fullflow project từ data → production:

### File structure:

```
end-to-end-project/
├── README.md              → Overview & Phase 1-3
├── DEPLOYMENT.md          → Phase 4-6 (Evaluation, API, .NET)
└── PIPELINE.md            → Phase 7-9 (Pipeline, Monitoring)
```

### Phases covered:

**Phase 1-3: Data & Model Development**

- ✅ Project structure setup
- ✅ Synthetic data generation
- ✅ Full preprocessing pipeline
- ✅ Multi-model training & comparison
- ✅ Hyperparameter tuning

**Phase 4-6: Deployment**

- ✅ Comprehensive evaluation module
- ✅ Flask REST API với endpoints:
  - `/predict` - Single prediction
  - `/predict_batch` - Batch predictions
  - `/health` - Health check
- ✅ .NET Core integration:
  - API Controller
  - Prediction Service
  - Models & DTOs
- ✅ Docker containerization

**Phase 7-9: Production & Maintenance**

- ✅ Automated pipeline script (`run_pipeline.py`)
- ✅ Performance monitoring
- ✅ Drift detection (performance & prediction)
- ✅ Automated retraining workflow
- ✅ Model versioning & backup
- ✅ Scheduled tasks (cron/Task Scheduler)

### Key Features:

- 🔄 **Full automation**: One command to run entire pipeline
- 📊 **Monitoring**: Track model performance over time
- 🔁 **Auto-retraining**: Detect drift and trigger retraining
- 🐍↔️.NET: **Python ML + .NET API** integration
- 🐳 **Docker**: Containerized deployment
- 📈 **Logging**: Complete audit trail
- 💾 **Backup**: Model versioning before retraining

---

## 🎯 Learning Path

### For Beginners:

1. Start with **[Preprocessing exercise](0-preprocessing/EXERCISE.md)**
2. Move to **[Regression exercise](1-regressions/EXERCISE.md)**
3. Practice each LESSON.md in folders 0-10
4. Build confidence with concepts

### For Intermediate:

1. Complete all topic exercises (1-3 above)
2. Try extension challenges in each exercise
3. Study **[End-to-End Project README](end-to-end-project/README.md)**
4. Implement Phases 1-3 yourself

### For Production-Ready:

1. Complete **end-to-end project** fully (all 9 phases)
2. Implement **DEPLOYMENT.md** with Docker
3. Set up **monitoring** system (PIPELINE.md)
4. Integrate with your own **.NET application**

---

## 💻 Quick Start

### Run an Exercise:

```bash
cd machine-learning/0-preprocessing
# Read EXERCISE.md
# Write your solution
# Compare with provided solution
```

### Run End-to-End Project:

```bash
cd machine-learning/end-to-end-project

# Create virtual environment
python -m venv venv
source venv/bin/activate  # Mac/Linux
# venv\Scripts\activate   # Windows

# Install dependencies
pip install -r requirements.txt

# Run full pipeline
python run_pipeline.py

# Start API
cd api
python app.py

# In another terminal, test
curl http://localhost:5000/health
```

### Setup .NET Client:

```bash
cd end-to-end-project/dotnet-client
dotnet new webapi -n ChurnPredictionAPI
# Copy provided code files
dotnet run

# Test .NET API
curl http://localhost:5001/api/prediction/example
```

---

## 📖 Documentation Structure

```
machine-learning/
├── STUDY_GUIDE.md                    ← Start here (overview)
├── EXERCISES.md                      ← This file
│
├── 0-preprocessing/
│   ├── LESSON.md                     ← Theory + syntax
│   └── EXERCISE.md                   ← Hands-on practice
│
├── 1-regressions/
│   ├── LESSON.md                     ← 6 regression algorithms
│   └── EXERCISE.md                   ← Compare models project
│
├── 2-classifications/
│   ├── LESSON.md                     ← 7 classification algorithms
│   └── EXERCISE.md                   ← (Coming soon)
│
... (3-10 similar structure)
│
└── end-to-end-project/
    ├── README.md                     ← Phases 1-3
    ├── DEPLOYMENT.md                 ← Phases 4-6
    ├── PIPELINE.md                   ← Phases 7-9
    ├── src/
    │   ├── data_preprocessing.py
    │   ├── train.py
    │   ├── evaluate.py
    │   ├── monitor.py
    │   └── retrain.py
    ├── api/
    │   └── app.py                    ← Flask API
    ├── dotnet-client/
    │   └── ChurnPredictionAPI/       ← .NET integration
    └── run_pipeline.py               ← Full automation
```

---

## 🎓 Tips for Success

1. **Hands-on first**: Code along with LESSON.md examples
2. **Then practice**: Do EXERCISE.md without looking at solution
3. **Compare**: Check your solution vs provided one
4. **Extend**: Try extension challenges
5. **Real project**: Adapt end-to-end project to your domain

---

## 🚀 After Completing

You will be able to:

- ✅ Process any tabular dataset
- ✅ Build & compare multiple ML models
- ✅ Deploy models as REST APIs
- ✅ Integrate ML into .NET applications
- ✅ Monitor & retrain models in production
- ✅ Build complete ML systems end-to-end

---

## 📞 Support

- Review [STUDY_GUIDE.md](STUDY_GUIDE.md) for topic overview
- Read LESSON.md for theory + syntax
- Do EXERCISE.md for practice
- Study end-to-end-project for production patterns

**Good luck! 🎉**
