# End-to-End Project - Part 3: Pipeline, Monitoring & Production

## 🔄 PHASE 7: Complete Pipeline Script

### 7.1 Full Pipeline Automation

**File: `run_pipeline.py`**

```python
import os
import sys
import subprocess
import json
from datetime import datetime

class MLPipeline:
    def __init__(self):
        self.start_time = datetime.now()
        self.log_file = f"logs/pipeline_{self.start_time.strftime('%Y%m%d_%H%M%S')}.log"
        os.makedirs("logs", exist_ok=True)

    def log(self, message):
        """Log message to console and file"""
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        log_message = f"[{timestamp}] {message}"
        print(log_message)
        with open(self.log_file, 'a') as f:
            f.write(log_message + '\n')

    def step_1_generate_data(self):
        """Generate synthetic dataset"""
        self.log("="*70)
        self.log("STEP 1: Generate Dataset")
        self.log("="*70)

        try:
            subprocess.run(['python', 'data/generate_data.py'], check=True)
            self.log("✅ Dataset generated successfully")
            return True
        except Exception as e:
            self.log(f"❌ Error: {e}")
            return False

    def step_2_preprocessing(self):
        """Data preprocessing"""
        self.log("\n" + "="*70)
        self.log("STEP 2: Data Preprocessing")
        self.log("="*70)

        try:
            subprocess.run(['python', 'src/data_preprocessing.py'], check=True)
            self.log("✅ Preprocessing complete")
            return True
        except Exception as e:
            self.log(f"❌ Error: {e}")
            return False

    def step_3_training(self):
        """Model training"""
        self.log("\n" + "="*70)
        self.log("STEP 3: Model Training")
        self.log("="*70)

        try:
            subprocess.run(['python', 'src/train.py'], check=True)
            self.log("✅ Training complete")
            return True
        except Exception as e:
            self.log(f"❌ Error: {e}")
            return False

    def step_4_evaluation(self):
        """Model evaluation"""
        self.log("\n" + "="*70)
        self.log("STEP 4: Model Evaluation")
        self.log("="*70)

        try:
            subprocess.run(['python', 'src/evaluate.py'], check=True)

            # Load and display metrics
            with open('models/evaluation_metrics.json', 'r') as f:
                metrics = json.load(f)

            self.log(f"\n📊 Final Metrics:")
            self.log(f"  Accuracy: {metrics['accuracy']:.4f}")
            self.log(f"  Precision: {metrics['precision']:.4f}")
            self.log(f"  Recall: {metrics['recall']:.4f}")
            self.log(f"  F1-Score: {metrics['f1']:.4f}")
            self.log(f"  ROC-AUC: {metrics['roc_auc']:.4f}")

            self.log("✅ Evaluation complete")
            return True
        except Exception as e:
            self.log(f"❌ Error: {e}")
            return False

    def step_5_validation(self):
        """Validate model meets production criteria"""
        self.log("\n" + "="*70)
        self.log("STEP 5: Production Validation")
        self.log("="*70)

        try:
            with open('models/evaluation_metrics.json', 'r') as f:
                metrics = json.load(f)

            # Production criteria
            MIN_ACCURACY = 0.80
            MIN_ROC_AUC = 0.75
            MIN_RECALL = 0.70  # Important for churn - don't miss churners

            checks = {
                'Accuracy >= 80%': metrics['accuracy'] >= MIN_ACCURACY,
                'ROC-AUC >= 75%': metrics['roc_auc'] >= MIN_ROC_AUC,
                'Recall >= 70%': metrics['recall'] >= MIN_RECALL
            }

            all_passed = True
            for check, passed in checks.items():
                status = "✅" if passed else "❌"
                self.log(f"  {status} {check}: {passed}")
                all_passed = all_passed and passed

            if all_passed:
                self.log("\n🎉 Model passed all production criteria!")
                return True
            else:
                self.log("\n⚠️  Model did not meet production criteria")
                return False

        except Exception as e:
            self.log(f"❌ Error: {e}")
            return False

    def step_6_deploy_prep(self):
        """Prepare for deployment"""
        self.log("\n" + "="*70)
        self.log("STEP 6: Deployment Preparation")
        self.log("="*70)

        # Create deployment package
        deployment_info = {
            'deployment_date': datetime.now().isoformat(),
            'model_path': 'models/best_model.pkl',
            'preprocessor_path': 'models/',
            'api_endpoint': 'http://localhost:5000/predict',
            'version': '1.0.0'
        }

        with open('models/deployment_info.json', 'w') as f:
            json.dump(deployment_info, f, indent=2)

        self.log("✅ Deployment package created")
        self.log(f"  Model: {deployment_info['model_path']}")
        self.log(f"  API: {deployment_info['api_endpoint']}")

        return True

    def run(self, skip_data_gen=False):
        """Run full pipeline"""
        self.log("🚀 Starting ML Pipeline")
        self.log(f"Start time: {self.start_time}")

        steps = []

        if not skip_data_gen:
            steps.append(('Generate Data', self.step_1_generate_data))

        steps.extend([
            ('Preprocessing', self.step_2_preprocessing),
            ('Training', self.step_3_training),
            ('Evaluation', self.step_4_evaluation),
            ('Validation', self.step_5_validation),
            ('Deployment Prep', self.step_6_deploy_prep)
        ])

        failed_steps = []

        for step_name, step_func in steps:
            if not step_func():
                failed_steps.append(step_name)
                self.log(f"\n❌ Pipeline failed at: {step_name}")
                break

        end_time = datetime.now()
        duration = (end_time - self.start_time).total_seconds()

        self.log("\n" + "="*70)
        self.log("PIPELINE SUMMARY")
        self.log("="*70)
        self.log(f"Start: {self.start_time}")
        self.log(f"End: {end_time}")
        self.log(f"Duration: {duration:.2f} seconds")

        if not failed_steps:
            self.log("\n✅ Pipeline completed successfully!")
            self.log(f"Log file: {self.log_file}")
            return True
        else:
            self.log(f"\n❌ Pipeline failed. Check {self.log_file} for details")
            return False

if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description='Run ML Pipeline')
    parser.add_argument('--skip-data-gen', action='store_true',
                       help='Skip data generation step')
    args = parser.parse_args()

    pipeline = MLPipeline()
    success = pipeline.run(skip_data_gen=args.skip_data_gen)

    sys.exit(0 if success else 1)
```

**Usage**:

```bash
# Full pipeline (including data generation)
python run_pipeline.py

# Skip data generation (if data already exists)
python run_pipeline.py --skip-data-gen
```

---

## 📈 PHASE 8: Monitoring & Retraining

### 8.1 Model Performance Monitoring

**File: `src/monitor.py`**

```python
import pandas as pd
import numpy as np
import joblib
import json
from datetime import datetime, timedelta
from sklearn.metrics import accuracy_score, roc_auc_score
import matplotlib.pyplot as plt

class ModelMonitor:
    def __init__(self, model_path='models/best_model.pkl'):
        self.model = joblib.load(model_path)
        self.performance_log = 'logs/performance_log.json'
        self.alert_threshold = {
            'accuracy_drop': 0.05,  # Alert if accuracy drops 5%
            'prediction_drift': 0.10  # Alert if prediction distribution changes 10%
        }

    def log_prediction(self, X, y_true=None, prediction_id=None):
        """Log prediction for monitoring"""

        y_pred = self.model.predict(X)
        y_proba = self.model.predict_proba(X)[:, 1]

        log_entry = {
            'timestamp': datetime.now().isoformat(),
            'prediction_id': prediction_id or str(datetime.now().timestamp()),
            'prediction': int(y_pred[0]) if len(y_pred) == 1 else y_pred.tolist(),
            'probability': float(y_proba[0]) if len(y_proba) == 1 else y_proba.tolist()
        }

        if y_true is not None:
            log_entry['actual'] = int(y_true[0]) if len(y_true) == 1 else y_true.tolist()

        # Append to log
        self._append_to_log(log_entry)

        return log_entry

    def _append_to_log(self, entry):
        """Append entry to performance log"""
        import os
        os.makedirs('logs', exist_ok=True)

        try:
            with open(self.performance_log, 'r') as f:
                logs = json.load(f)
        except FileNotFoundError:
            logs = []

        logs.append(entry)

        with open(self.performance_log, 'w') as f:
            json.dump(logs, f)

    def check_performance_drift(self, baseline_metrics_path='models/evaluation_metrics.json'):
        """Check if model performance has drifted"""

        # Load baseline metrics
        with open(baseline_metrics_path, 'r') as f:
            baseline = json.load(f)

        # Load recent predictions
        try:
            with open(self.performance_log, 'r') as f:
                logs = json.load(f)
        except FileNotFoundError:
            print("No performance logs found")
            return

        # Filter logs with actual values (last 30 days)
        cutoff_date = datetime.now() - timedelta(days=30)
        recent_logs = [
            log for log in logs
            if 'actual' in log and
            datetime.fromisoformat(log['timestamp']) > cutoff_date
        ]

        if len(recent_logs) < 100:
            print(f"Insufficient data for drift detection ({len(recent_logs)} samples)")
            return

        # Calculate current performance
        y_true = np.array([log['actual'] for log in recent_logs])
        y_pred = np.array([log['prediction'] for log in recent_logs])
        y_proba = np.array([log['probability'] for log in recent_logs])

        current_accuracy = accuracy_score(y_true, y_pred)
        current_roc_auc = roc_auc_score(y_true, y_proba)

        # Compare with baseline
        accuracy_drop = baseline['accuracy'] - current_accuracy
        roc_drop = baseline['roc_auc'] - current_roc_auc

        print("="*60)
        print("PERFORMANCE DRIFT CHECK")
        print("="*60)
        print(f"Samples analyzed: {len(recent_logs)}")
        print(f"\nBaseline Accuracy: {baseline['accuracy']:.4f}")
        print(f"Current Accuracy: {current_accuracy:.4f}")
        print(f"Drop: {accuracy_drop:.4f} ({accuracy_drop/baseline['accuracy']*100:.1f}%)")

        print(f"\nBaseline ROC-AUC: {baseline['roc_auc']:.4f}")
        print(f"Current ROC-AUC: {current_roc_auc:.4f}")
        print(f"Drop: {roc_drop:.4f} ({roc_drop/baseline['roc_auc']*100:.1f}%)")

        # Alert if significant drift
        if accuracy_drop > self.alert_threshold['accuracy_drop']:
            print(f"\n⚠️  ALERT: Accuracy dropped by {accuracy_drop:.2%}!")
            print("   Recommend retraining the model")
            return True  # Needs retraining
        else:
            print("\n✅ Model performance is stable")
            return False

    def check_prediction_drift(self):
        """Check if prediction distribution has changed"""

        try:
            with open(self.performance_log, 'r') as f:
                logs = json.load(f)
        except FileNotFoundError:
            print("No performance logs found")
            return

        # Split into old (30-60 days ago) and recent (last 30 days)
        now = datetime.now()
        recent_cutoff = now - timedelta(days=30)
        old_cutoff = now - timedelta(days=60)

        recent_logs = [log for log in logs if datetime.fromisoformat(log['timestamp']) > recent_cutoff]
        old_logs = [log for log in logs
                   if old_cutoff < datetime.fromisoformat(log['timestamp']) <= recent_cutoff]

        if len(recent_logs) < 50 or len(old_logs) < 50:
            print("Insufficient data for drift detection")
            return

        # Compare prediction distributions
        recent_preds = np.array([log['prediction'] for log in recent_logs])
        old_preds = np.array([log['prediction'] for log in old_logs])

        recent_churn_rate = recent_preds.mean()
        old_churn_rate = old_preds.mean()

        drift = abs(recent_churn_rate - old_churn_rate)

        print("="*60)
        print("PREDICTION DRIFT CHECK")
        print("="*60)
        print(f"Old period (30-60 days ago): {len(old_logs)} predictions")
        print(f"  Churn rate: {old_churn_rate:.2%}")
        print(f"\nRecent period (last 30 days): {len(recent_logs)} predictions")
        print(f"  Churn rate: {recent_churn_rate:.2%}")
        print(f"\nDrift: {drift:.2%}")

        if drift > self.alert_threshold['prediction_drift']:
            print(f"\n⚠️  ALERT: Prediction distribution has shifted!")
            print("   This could indicate data drift")
            return True
        else:
            print("\n✅ Prediction distribution is stable")
            return False

# Usage
if __name__ == '__main__':
    monitor = ModelMonitor()

    # Check for performance drift
    performance_drifted = monitor.check_performance_drift()

    # Check for prediction drift
    prediction_drifted = monitor.check_prediction_drift()

    if performance_drifted or prediction_drifted:
        print("\n🔄 RECOMMENDATION: Retrain the model")
    else:
        print("\n✅ No action needed")
```

### 8.2 Automated Retraining Script

**File: `src/retrain.py`**

```python
import pandas as pd
import numpy as np
from datetime import datetime
import joblib
import json
import shutil
import os

def backup_current_model():
    """Backup current model before retraining"""
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    backup_dir = f'models/backups/backup_{timestamp}'

    os.makedirs(backup_dir, exist_ok=True)

    # Backup all model files
    files_to_backup = [
        'best_model.pkl',
        'scaler.pkl',
        'label_encoders.pkl',
        'onehot_encoder.pkl',
        'imputer.pkl',
        'model_metadata.json',
        'evaluation_metrics.json',
        'feature_names.json'
    ]

    for file in files_to_backup:
        src = f'models/{file}'
        if os.path.exists(src):
            shutil.copy(src, backup_dir)

    print(f"✅ Model backed up to {backup_dir}")
    return backup_dir

def retrain_model(new_data_path=None):
    """Retrain model with new data"""

    print("="*70)
    print("MODEL RETRAINING")
    print("="*70)

    # 1. Backup current model
    backup_path = backup_current_model()

    # 2. Load new data (or use existing)
    if new_data_path:
        print(f"\nLoading new data from {new_data_path}")
        df = pd.read_csv(new_data_path)
    else:
        print("\nUsing existing data")
        df = pd.read_csv('data/raw/telecom_churn.csv')

    # 3. Run full pipeline
    from data_preprocessing import DataPreprocessor
    from train import ModelTrainer
    from evaluate import ModelEvaluator

    # Preprocessing
    print("\n1. Preprocessing...")
    preprocessor = DataPreprocessor()
    X_train, X_test, y_train, y_test = preprocessor.fit_transform(df)

    # Save preprocessed data
    np.save('data/processed/X_train.npy', X_train)
    np.save('data/processed/X_test.npy', X_test)
    np.save('data/processed/y_train.npy', y_train)
    np.save('data/processed/y_test.npy', y_test)
    preprocessor.save('models/')

    # Training
    print("\n2. Training models...")
    trainer = ModelTrainer()
    trainer.train_multiple_models(X_train, y_train, X_test, y_test)
    trainer.save_model()

    # Evaluation
    print("\n3. Evaluating...")
    evaluator = ModelEvaluator('models/best_model.pkl')
    new_metrics = evaluator.comprehensive_evaluation(X_test, y_test)

    # 4. Compare with old model
    try:
        with open(f'{backup_path}/evaluation_metrics.json', 'r') as f:
            old_metrics = json.load(f)

        print("\n" + "="*70)
        print("RETRAINING COMPARISON")
        print("="*70)
        print(f"{'Metric':<20} {'Old Model':<15} {'New Model':<15} {'Change':<15}")
        print("-"*70)

        for metric in ['accuracy', 'precision', 'recall', 'f1', 'roc_auc']:
            old_val = old_metrics[metric]
            new_val = new_metrics[metric]
            change = new_val - old_val
            change_pct = (change / old_val * 100) if old_val > 0 else 0

            status = "↑" if change > 0 else ("↓" if change < 0 else "→")
            print(f"{metric:<20} {old_val:<15.4f} {new_val:<15.4f} {status} {change_pct:>6.2f}%")

        # Decision: Keep new model or rollback?
        if new_metrics['roc_auc'] < old_metrics['roc_auc'] - 0.02:  # 2% worse
            print("\n⚠️  New model performs worse. Consider rollback.")
            print(f"   To rollback: cp {backup_path}/* models/")
        else:
            print("\n✅ New model accepted!")

    except:
        print("\n✅ Retraining complete (no comparison available)")

    return new_metrics

if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description='Retrain model')
    parser.add_argument('--data', type=str, help='Path to new training data CSV')
    args = parser.parse_args()

    retrain_model(args.data)
```

### 8.3 Scheduled Retraining (Cron/Task Scheduler)

**File: `scripts/schedule_retrain.sh`** (Linux/Mac)

```bash
#!/bin/bash

# Add to crontab: crontab -e
# Run retraining every Sunday at 2 AM:
# 0 2 * * 0 /path/to/schedule_retrain.sh

cd /path/to/end-to-end-project

# Activate virtual environment
source venv/bin/activate

# Run monitoring
python src/monitor.py > logs/monitor_$(date +\%Y\%m\%d).log 2>&1

# If drift detected, retrain
if [ $? -eq 1 ]; then
    echo "Drift detected, starting retraining..."
    python src/retrain.py >> logs/retrain_$(date +\%Y\%m\%d).log 2>&1

    # Restart API
    sudo systemctl restart churn-api
fi
```

**Windows Task Scheduler** (`scripts/schedule_retrain.ps1`):

```powershell
# Schedule with Task Scheduler
# Task Scheduler > Create Task > Triggers: Weekly, Sunday 2:00 AM
# Actions: Run PowerShell script

cd C:\path\to\end-to-end-project

# Activate venv
.\.venv\Scripts\Activate.ps1

# Run monitoring
python src\monitor.py | Out-File -FilePath "logs\monitor_$(Get-Date -Format 'yyyyMMdd').log"

# Check exit code and retrain if needed
if ($LASTEXITCODE -eq 1) {
    Write-Host "Drift detected, starting retraining..."
    python src\retrain.py | Out-File -Append -FilePath "logs\retrain_$(Get-Date -Format 'yyyyMMdd').log"

    # Restart API service
    Restart-Service ChurnPredictionAPI
}
```

---

## 🎯 PHASE 9: Quick Start Guide

### 9.1 First Time Setup

```bash
# 1. Clone/create project
mkdir end-to-end-project
cd end-to-end-project

# 2. Create virtual environment
python -m venv venv
source venv/bin/activate  # Linux/Mac
# venv\Scripts\activate  # Windows

# 3. Install dependencies
pip install -r requirements.txt

# 4. Run full pipeline
python run_pipeline.py

# 5. Start Python API
cd api
python app.py

# 6. Test API
curl -X POST http://localhost:5000/predict \
  -H "Content-Type: application/json" \
  -d '{"Age": 35, "Gender": "Male", "Tenure": 24, ...}'
```

### 9.2 .NET Client Setup

```bash
# 1. Create .NET project
cd dotnet-client
dotnet new webapi -n ChurnPredictionAPI

# 2. Add packages
dotnet add package Microsoft.AspNetCore.OpenApi
dotnet add package Swashbuckle.AspNetCore

# 3. Copy code files (controllers, services, models)

# 4. Run
dotnet run

# 5. Test
curl http://localhost:5001/api/prediction/example
```

---

## 📦 Complete requirements.txt

**File: `requirements.txt`**

```
pandas==2.0.3
numpy==1.24.3
scikit-learn==1.3.0
xgboost==1.7.6
matplotlib==3.7.2
seaborn==0.12.2
joblib==1.3.0
flask==2.3.0
flask-cors==4.0.0
```

---

##✅ Summary: Complete Workflow

1. **Data Generation**: `python data/generate_data.py`
2. **Full Pipeline**: `python run_pipeline.py`
3. **Start API**: `python api/app.py`
4. **.NET Integration**: `dotnet run` in dotnet-client
5. **Monitoring**: `python src/monitor.py` (schedule daily/weekly)
6. **Retraining**: `python src/retrain.py` (when drift detected)

**Production Checklist**:

- ✅ Model accuracy > 80%
- ✅ API health check working
- ✅ .NET client can call Python API
- ✅ Monitoring scheduled
- ✅ Backup strategy in place
- ✅ Logging configured
