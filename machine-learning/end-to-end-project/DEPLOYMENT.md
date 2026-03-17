# End-to-End Project - Part 2: Evaluation, Deployment & Production

## 📊 PHASE 4: Model Evaluation

### 4.1 Evaluation Module

**File: `src/evaluate.py`**

```python
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.metrics import (
    classification_report, confusion_matrix,
    roc_auc_score, roc_curve, precision_recall_curve,
    accuracy_score, precision_score, recall_score, f1_score
)
import joblib

class ModelEvaluator:
    def __init__(self, model_path='models/best_model.pkl'):
        self.model = joblib.load(model_path)

    def comprehensive_evaluation(self, X_test, y_test, save_plots=True):
        """Complete evaluation with metrics and visualizations"""

        # Predictions
        y_pred = self.model.predict(X_test)
        y_pred_proba = self.model.predict_proba(X_test)[:, 1]

        print("="*70)
        print("MODEL EVALUATION REPORT")
        print("="*70)

        # 1. Basic Metrics
        print("\n📊 Classification Metrics:")
        print(f"  Accuracy: {accuracy_score(y_test, y_pred):.4f}")
        print(f"  Precision: {precision_score(y_test, y_pred):.4f}")
        print(f"  Recall: {recall_score(y_test, y_pred):.4f}")
        print(f"  F1-Score: {f1_score(y_test, y_pred):.4f}")
        print(f"  ROC-AUC: {roc_auc_score(y_test, y_pred_proba):.4f}")

        # 2. Classification Report
        print("\n📋 Detailed Classification Report:")
        print(classification_report(y_test, y_pred, target_names=['Not Churn', 'Churn']))

        # 3. Confusion Matrix
        cm = confusion_matrix(y_test, y_pred)
        print("\n🎯 Confusion Matrix:")
        print(cm)
        print(f"\nTrue Negatives: {cm[0,0]}")
        print(f"False Positives: {cm[0,1]}")
        print(f"False Negatives: {cm[1,0]}")
        print(f"True Positives: {cm[1,1]}")

        # Business metrics
        total = cm.sum()
        correct_predictions = cm[0,0] + cm[1,1]
        false_alarms = cm[0,1]  # Predicted churn but didn't
        missed_churns = cm[1,0]  # Didn't predict but churned

        print(f"\n💼 Business Metrics:")
        print(f"  Total Predictions: {total}")
        print(f"  Correct Predictions: {correct_predictions} ({correct_predictions/total*100:.1f}%)")
        print(f"  False Alarms (unnecessary retention effort): {false_alarms} ({false_alarms/total*100:.1f}%)")
        print(f"  Missed Churns (lost customers): {missed_churns} ({missed_churns/total*100:.1f}%)")

        if save_plots:
            self.plot_evaluation(y_test, y_pred, y_pred_proba)

        return {
            'accuracy': accuracy_score(y_test, y_pred),
            'precision': precision_score(y_test, y_pred),
            'recall': recall_score(y_test, y_pred),
            'f1': f1_score(y_test, y_pred),
            'roc_auc': roc_auc_score(y_test, y_pred_proba),
            'confusion_matrix': cm.tolist()
        }

    def plot_evaluation(self, y_test, y_pred, y_pred_proba):
        """Create evaluation visualizations"""

        fig, axes = plt.subplots(2, 2, figsize=(15, 12))

        # 1. Confusion Matrix
        cm = confusion_matrix(y_test, y_pred)
        sns.heatmap(cm, annot=True, fmt='d', cmap='Blues', ax=axes[0,0])
        axes[0,0].set_title('Confusion Matrix')
        axes[0,0].set_xlabel('Predicted')
        axes[0,0].set_ylabel('Actual')
        axes[0,0].set_xticklabels(['Not Churn', 'Churn'])
        axes[0,0].set_yticklabels(['Not Churn', 'Churn'])

        # 2. ROC Curve
        fpr, tpr, thresholds = roc_curve(y_test, y_pred_proba)
        roc_auc = roc_auc_score(y_test, y_pred_proba)

        axes[0,1].plot(fpr, tpr, label=f'ROC Curve (AUC = {roc_auc:.3f})', linewidth=2)
        axes[0,1].plot([0, 1], [0, 1], 'k--', label='Random Classifier')
        axes[0,1].set_xlabel('False Positive Rate')
        axes[0,1].set_ylabel('True Positive Rate')
        axes[0,1].set_title('ROC Curve')
        axes[0,1].legend()
        axes[0,1].grid(True)

        # 3. Precision-Recall Curve
        precision, recall, _ = precision_recall_curve(y_test, y_pred_proba)

        axes[1,0].plot(recall, precision, linewidth=2)
        axes[1,0].set_xlabel('Recall')
        axes[1,0].set_ylabel('Precision')
        axes[1,0].set_title('Precision-Recall Curve')
        axes[1,0].grid(True)

        # 4. Prediction Distribution
        axes[1,1].hist(y_pred_proba[y_test==0], bins=50, alpha=0.5, label='Not Churn', color='blue')
        axes[1,1].hist(y_pred_proba[y_test==1], bins=50, alpha=0.5, label='Churn', color='red')
        axes[1,1].set_xlabel('Predicted Probability')
        axes[1,1].set_ylabel('Frequency')
        axes[1,1].set_title('Probability Distribution by True Class')
        axes[1,1].legend()
        axes[1,1].grid(True)

        plt.tight_layout()
        plt.savefig('models/evaluation_plots.png', dpi=300, bbox_inches='tight')
        print("\n✅ Evaluation plots saved to 'models/evaluation_plots.png'")
        plt.close()

    def analyze_feature_importance(self, feature_names=None):
        """Analyze and plot feature importance"""

        if not hasattr(self.model, 'feature_importances_'):
            print("Model doesn't support feature importance")
            return

        importances = self.model.feature_importances_

        if feature_names is None:
            feature_names = [f'Feature {i}' for i in range(len(importances))]

        # Sort by importance
        indices = np.argsort(importances)[::-1]

        print("\n🔍 Top 10 Most Important Features:")
        for i in range(min(10, len(importances))):
            idx = indices[i]
            print(f"  {i+1}. {feature_names[idx]}: {importances[idx]:.4f}")

        # Plot
        plt.figure(figsize=(10, 6))
        top_n = 15
        top_indices = indices[:top_n]
        plt.barh(range(top_n), importances[top_indices])
        plt.yticks(range(top_n), [feature_names[i] for i in top_indices])
        plt.xlabel('Importance')
        plt.title(f'Top {top_n} Feature Importances')
        plt.tight_layout()
        plt.savefig('models/feature_importance.png', dpi=300, bbox_inches='tight')
        print("✅ Feature importance plot saved to 'models/feature_importance.png'")
        plt.close()

# Usage
if __name__ == '__main__':
    # Load test data
    X_test = np.load('data/processed/X_test.npy')
    y_test = np.load('data/processed/y_test.npy')

    # Evaluate
    evaluator = ModelEvaluator('models/best_model.pkl')
    metrics = evaluator.comprehensive_evaluation(X_test, y_test)

    # Feature importance (if applicable)
    try:
        import json
        with open('models/feature_names.json', 'r') as f:
            feature_names = json.load(f)
        evaluator.analyze_feature_importance(feature_names)
    except:
        print("\nSkipping feature importance analysis")

    # Save metrics
    import json
    with open('models/evaluation_metrics.json', 'w') as f:
        json.dump(metrics, f, indent=2)

    print("\n✅ Evaluation complete!")
```

---

## 🚀 PHASE 5: Deployment - Python Flask API

### 5.1 Flask API

**File: `api/app.py`**

```python
from flask import Flask, request, jsonify
from flask_cors import CORS
import joblib
import numpy as np
import pandas as pd
import sys
sys.path.append('..')
from src.data_preprocessing import DataPreprocessor

app = Flask(__name__)
CORS(app)  # Enable CORS for .NET client

# Load model and preprocessor at startup
print("Loading model and preprocessor...")
model = joblib.load('../models/best_model.pkl')
preprocessor = DataPreprocessor.load('../models/')
print("✅ Model loaded successfully!")

@app.route('/')
def home():
    return jsonify({
        'message': 'Churn Prediction API',
        'version': '1.0',
        'endpoints': {
            '/predict': 'POST - Predict churn for single customer',
            '/predict_batch': 'POST - Predict churn for multiple customers',
            '/health': 'GET - Health check'
        }
    })

@app.route('/health')
def health():
    return jsonify({'status': 'healthy', 'model_loaded': model is not None})

@app.route('/predict', methods=['POST'])
def predict():
    """
    Predict churn for a single customer

    Request JSON example:
    {
        "Age": 35,
        "Gender": "Male",
        "Tenure": 24,
        "MonthlyCharges": 79.99,
        "TotalCharges": 1919.76,
        "Contract": "One year",
        "InternetService": "Fiber optic",
        "TechSupport": "Yes",
        "OnlineSecurity": "No",
        "PaymentMethod": "Electronic check"
    }
    """
    try:
        data = request.json

        # Convert to DataFrame
        df = pd.DataFrame([data])

        # Preprocess
        X = preprocessor.transform(df)

        # Predict
        prediction = model.predict(X)[0]
        probability = model.predict_proba(X)[0]

        # Prepare response
        response = {
            'churn_prediction': int(prediction),
            'churn_probability': float(probability[1]),
            'confidence': float(max(probability)),
            'risk_level': 'High' if probability[1] > 0.7 else ('Medium' if probability[1] > 0.4 else 'Low'),
            'details': {
                'probability_not_churn': float(probability[0]),
                'probability_churn': float(probability[1])
            }
        }

        return jsonify(response), 200

    except Exception as e:
        return jsonify({'error': str(e)}), 400

@app.route('/predict_batch', methods=['POST'])
def predict_batch():
    """
    Predict churn for multiple customers

    Request JSON example:
    {
        "customers": [
            {"Age": 35, "Gender": "Male", ...},
            {"Age": 42, "Gender": "Female", ...}
        ]
    }
    """
    try:
        data = request.json
        customers = data['customers']

        # Convert to DataFrame
        df = pd.DataFrame(customers)

        # Preprocess
        X = preprocessor.transform(df)

        # Predict
        predictions = model.predict(X)
        probabilities = model.predict_proba(X)

        # Prepare response
        results = []
        for i, (pred, proba) in enumerate(zip(predictions, probabilities)):
            results.append({
                'customer_index': i,
                'churn_prediction': int(pred),
                'churn_probability': float(proba[1]),
                'risk_level': 'High' if proba[1] > 0.7 else ('Medium' if proba[1] > 0.4 else 'Low')
            })

        response = {
            'total_customers': len(results),
            'predicted_churns': int(sum(predictions)),
            'predictions': results
        }

        return jsonify(response), 200

    except Exception as e:
        return jsonify({'error': str(e)}), 400

@app.route('/model_info')
def model_info():
    """Get model metadata"""
    try:
        import json
        with open('../models/model_metadata.json', 'r') as f:
            metadata = json.load(f)
        return jsonify(metadata), 200
    except:
        return jsonify({'error': 'Metadata not available'}), 404

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
```

### 5.2 Requirements File

**File: `api/requirements.txt`**

```
flask==2.3.0
flask-cors==4.0.0
scikit-learn==1.3.0
pandas==2.0.3
numpy==1.24.3
xgboost==1.7.6
joblib==1.3.0
```

### 5.3 Dockerfile

**File: `api/Dockerfile`**

```dockerfile
FROM python:3.9-slim

WORKDIR /app

# Copy requirements
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY app.py .
COPY ../models ./models
COPY ../src ./src

EXPOSE 5000

CMD ["python", "app.py"]
```

### 5.4 Run API

```bash
# Install dependencies
cd api
pip install -r requirements.txt

# Run Flask API
python app.py

# Test with curl
curl -X POST http://localhost:5000/predict \
  -H "Content-Type: application/json" \
  -d '{
    "Age": 35,
    "Gender": "Male",
    "Tenure": 24,
    "MonthlyCharges": 79.99,
    "TotalCharges": 1919.76,
    "Contract": "One year",
    "InternetService": "Fiber optic",
    "TechSupport": "Yes",
    "OnlineSecurity": "No",
    "PaymentMethod": "Electronic check"
  }'
```

---

## 💻 PHASE 6: .NET Integration

### 6.1 .NET Models

**File: `dotnet-client/ChurnPredictionAPI/Models/ChurnRequest.cs`**

```csharp
namespace ChurnPredictionAPI.Models
{
    public class ChurnRequest
    {
        public int Age { get; set; }
        public string Gender { get; set; }
        public int Tenure { get; set; }
        public decimal MonthlyCharges { get; set; }
        public decimal TotalCharges { get; set; }
        public string Contract { get; set; }
        public string InternetService { get; set; }
        public string TechSupport { get; set; }
        public string OnlineSecurity { get; set; }
        public string PaymentMethod { get; set; }
    }

    public class ChurnResponse
    {
        public int ChurnPrediction { get; set; }
        public double ChurnProbability { get; set; }
        public double Confidence { get; set; }
        public string RiskLevel { get; set; }
        public ChurnDetails Details { get; set; }
    }

    public class ChurnDetails
    {
        public double ProbabilityNotChurn { get; set; }
        public double ProbabilityChurn { get; set; }
    }
}
```

### 6.2 Prediction Service

**File: `dotnet-client/ChurnPredictionAPI/Services/MLPredictionService.cs`**

```csharp
using System.Net.Http.Json;
using ChurnPredictionAPI.Models;

namespace ChurnPredictionAPI.Services
{
    public class MLPredictionService
    {
        private readonly HttpClient _httpClient;
        private readonly ILogger<MLPredictionService> _logger;
        private readonly string _pythonApiUrl;

        public MLPredictionService(
            HttpClient httpClient,
            IConfiguration configuration,
            ILogger<MLPredictionService> logger)
        {
            _httpClient = httpClient;
            _logger = logger;
            _pythonApiUrl = configuration["PythonAPI:BaseUrl"] ?? "http://localhost:5000";
        }

        public async Task<ChurnResponse?> PredictChurnAsync(ChurnRequest request)
        {
            try
            {
                _logger.LogInformation("Calling Python ML API for prediction");

                var response = await _httpClient.PostAsJsonAsync(
                    $"{_pythonApiUrl}/predict",
                    request
                );

                response.EnsureSuccessStatusCode();

                var result = await response.Content.ReadFromJsonAsync<ChurnResponse>();

                _logger.LogInformation(
                    "Prediction received: ChurnProbability={Probability}, RiskLevel={RiskLevel}",
                    result?.ChurnProbability,
                    result?.RiskLevel
                );

                return result;
            }
            catch (HttpRequestException ex)
            {
                _logger.LogError(ex, "Error calling Python ML API");
                throw new Exception("Failed to get prediction from ML service", ex);
            }
        }

        public async Task<bool> HealthCheckAsync()
        {
            try
            {
                var response = await _httpClient.GetAsync($"{_pythonApiUrl}/health");
                return response.IsSuccessStatusCode;
            }
            catch
            {
                return false;
            }
        }
    }
}
```

### 6.3 API Controller

**File: `dotnet-client/ChurnPredictionAPI/Controllers/PredictionController.cs`**

```csharp
using Microsoft.AspNetCore.Mvc;
using ChurnPredictionAPI.Models;
using ChurnPredictionAPI.Services;

namespace ChurnPredictionAPI.Controllers
{
    [ApiController]
    [Route("api/[controller]")]
    public class PredictionController : ControllerBase
    {
        private readonly MLPredictionService _predictionService;
        private readonly ILogger<PredictionController> _logger;

        public PredictionController(
            MLPredictionService predictionService,
            ILogger<PredictionController> logger)
        {
            _predictionService = predictionService;
            _logger = logger;
        }

        [HttpPost("churn")]
        public async Task<IActionResult> PredictChurn([FromBody] ChurnRequest request)
        {
            if (!ModelState.IsValid)
            {
                return BadRequest(ModelState);
            }

            try
            {
                var prediction = await _predictionService.PredictChurnAsync(request);

                if (prediction == null)
                {
                    return StatusCode(500, "Failed to get prediction");
                }

                return Ok(prediction);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in PredictChurn endpoint");
                return StatusCode(500, new { error = ex.Message });
            }
        }

        [HttpGet("health")]
        public async Task<IActionResult> HealthCheck()
        {
            var isHealthy = await _predictionService.HealthCheckAsync();

            if (isHealthy)
            {
                return Ok(new { status = "healthy", mlService = "connected" });
            }

            return StatusCode(503, new { status = "unhealthy", mlService = "disconnected" });
        }

        [HttpPost("example")]
        public async Task<IActionResult> GetExamplePrediction()
        {
            // Example customer data
            var exampleRequest = new ChurnRequest
            {
                Age = 35,
                Gender = "Male",
                Tenure = 24,
                MonthlyCharges = 79.99m,
                TotalCharges = 1919.76m,
                Contract = "One year",
                InternetService = "Fiber optic",
                TechSupport = "Yes",
                OnlineSecurity = "No",
                PaymentMethod = "Electronic check"
            };

            var prediction = await _predictionService.PredictChurnAsync(exampleRequest);

            return Ok(new
            {
                request = exampleRequest,
                prediction
            });
        }
    }
}
```

### 6.4 Program.cs

**File: `dotnet-client/ChurnPredictionAPI/Program.cs`**

```csharp
using ChurnPredictionAPI.Services;

var builder = WebApplication.CreateBuilder(args);

// Add services
builder.Services.AddControllers();
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

// Configure HttpClient for ML service
builder.Services.AddHttpClient<MLPredictionService>();

// Add CORS
builder.Services.AddCors(options =>
{
    options.AddPolicy("AllowAll", policy =>
    {
        policy.AllowAnyOrigin()
              .AllowAnyMethod()
              .AllowAnyHeader();
    });
});

var app = builder.Build();

// Configure pipeline
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseCors("AllowAll");
app.UseAuthorization();
app.MapControllers();

app.Run();
```

### 6.5 appsettings.json

**File: `dotnet-client/ChurnPredictionAPI/appsettings.json`**

```json
{
  "Logging": {
    "LogLevel": {
      "Default": "Information",
      "Microsoft.AspNetCore": "Warning"
    }
  },
  "AllowedHosts": "*",
  "PythonAPI": {
    "BaseUrl": "http://localhost:5000"
  }
}
```

---

_Tiếp tục phần 3 với Full Pipeline Script, Monitoring & Retraining..._
