using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Text;
using System.Xml.Linq;

public partial class MainViewModel : ObservableObject
{
    public ICommand RunNomadListCommand { get; }

    // Collection pour stocker les applications récupérées
    public ObservableCollection<string> Applications { get; } = new ObservableCollection<string>();

    public MainViewModel()
    {
      
        RunNomadListCommand = new RelayCommand(RunNomadList);
    }

    private async void RunNomadList()
    {
        string exePath = System.IO.Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tools", "nomad.exe");
        Debug.WriteLine($"Executable Path: {exePath}");

        Process process = new Process();
        process.StartInfo.FileName = exePath;
        process.StartInfo.Arguments = "list";
        process.StartInfo.WorkingDirectory = System.IO.Path.GetDirectoryName(exePath);
        process.StartInfo.RedirectStandardOutput = true;
        process.StartInfo.RedirectStandardError = true;
        process.StartInfo.UseShellExecute = false;
        process.StartInfo.CreateNoWindow = true;

        process.StartInfo.StandardOutputEncoding = Encoding.UTF8;
        process.StartInfo.StandardErrorEncoding = Encoding.UTF8;

        try
        {
            bool isStarted = process.Start();
            if (!isStarted)
            {
                name = "Failed to start the process.";
                return;
            }

            var outputTask = process.StandardOutput.ReadToEndAsync();
            var errorTask = process.StandardError.ReadToEndAsync();

            await Task.WhenAll(outputTask, errorTask);

            process.WaitForExit();

            string standardOutput = outputTask.Result;
            string standardError = errorTask.Result;

            string allAppsStr = ((standardError + standardOutput).Split(":")).Last();
            string[] apps = allAppsStr.Split(",");

            Applications.Clear();  // Vider la collection avant d'ajouter les nouvelles apps
            foreach (var app in apps)
            {
                string trimmedApp = app.Trim();
                if (!string.IsNullOrEmpty(trimmedApp))
                {
                    Applications.Add(trimmedApp);
                }
            }
        }
        catch (Exception ex)
        {
            Name = $"Exception: {ex.Message}";
        }
    }

    [ObservableProperty]
    private string name;
}
